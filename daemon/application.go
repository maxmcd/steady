package daemon

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "embed"

	"github.com/benbjohnson/litestream"
	"github.com/maxmcd/steady/internal/execx"
	"github.com/maxmcd/steady/internal/ring"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/internal/syncx"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/exp/slog"
)

var (
	maxBufferedRequests = 10
)

//go:embed wrapper.ts
var wrapperTS []byte

// application is an application instance. Data for the application is stored in
// a directory and the application is started to handle requests and killed
// after a period of inactivity.
type application struct {
	name string

	// dir is the directory that contains all application data
	dir  string
	port int

	DBs []string
	Env []string

	dbLimitWaiter *syncx.LimitedBroadcast

	mutex           sync.Mutex
	inFlightCounter int

	cmd     *execx.Cmd
	running bool

	stopRequestChan chan struct{}
	resetKillTimer  chan struct{}

	requestCount int
	startCount   int

	cancel func()

	dbs          map[string]*litestream.DB
	createDBFunc func(string) (*litestream.DB, error)
}

func (d *Daemon) newApplication(name string, dir string, port int) *application {
	a := &application{
		name:            name,
		dir:             dir,
		port:            port,
		stopRequestChan: make(chan struct{}),
		resetKillTimer:  make(chan struct{}),
		createDBFunc:    d.createDB(name),
		dbLimitWaiter:   syncx.NewLimitedBroadcast(maxBufferedRequests),
		dbs:             make(map[string]*litestream.DB),
	}
	return a
}
func (a *application) waitForDB() {
	a.dbLimitWaiter.StartWait()
}

func (a *application) dbDownloaded() error {
	if err := a.checkForDBs(); err != nil {
		return err
	}
	a.dbLimitWaiter.Signal()
	return nil
}

func (a *application) start() {
	go a.runLoop()
}

func (a *application) runLoop() {
	a.mutex.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.mutex.Unlock()
	killTimer := time.NewTimer(math.MaxInt64)
	for {
		select {
		case <-a.resetKillTimer:
			killTimer.Reset(math.MaxInt64)
		case <-a.stopRequestChan:
			killTimer.Reset(time.Second)
		case <-killTimer.C:
			a.stopProcess(false)
		case <-ctx.Done():
			return
		}
	}
}

func (a *application) updateApplication(src []byte) error {
	// Queue future requests
	a.dbLimitWaiter.StartWait()
	// TODO: drain requests
	a.stopProcess(true)
	a.resetKillTimer <- struct{}{}
	defer a.dbLimitWaiter.Signal()

	f, err := os.Create(filepath.Join(a.dir, "index.ts"))
	if err != nil {
		return err
	}

	if _, err := f.Write(src); err != nil {
		return err
	}
	// TODO: ensure it is healthy, roll back if needed
	return nil
}

func (a *application) stopProcess(force bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// TODO: audit this, if we try to stop and in-flight counter is >0 do we
	// always shut down correctly?
	if !force && a.inFlightCounter > 0 {
		return
	}
	if !a.running {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	if err := a.cmd.Shutdown(ctx); err != nil {
		slog.Error("error stopping process", err)
	}
	cancel()
	a.running = false
	if err := a.checkForDBs(); err != nil {
		slog.Error("error checking for dbs", err)
	}
}

// shutdown will completely shut down all applications and clean up
func (a *application) shutdown() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cancel()
	var stopErr error
	if a.running {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		stopErr = a.cmd.Shutdown(ctx)
		cancel()
		a.running = false
	}
	var err error
	if a.dbs != nil {
		for _, db := range a.dbs {
			if e := db.Close(); e != nil && err == nil {
				err = fmt.Errorf("close db: path=%s err=%w", db.Path(), e)
			}
		}
	}
	if stopErr != nil {
		return stopErr
	}
	return err
}

func (a *application) checkForDBs() error {
	dbPaths, err := filepath.Glob(filepath.Join(a.dir, "./*.sqlite"))
	if err != nil {
		return err
	}
	foundDBFiles := map[string]struct{}{}
	for _, dbPath := range dbPaths {
		foundDBFiles[dbPath] = struct{}{}
	}

	for _, db := range a.dbs {
		path := db.Path()
		// If we didn't find a litestream db in our project, remove it.
		// TODO: Figure out when this will happen, if litestream will event allow it,
		// other cleanup we have to do, what is the expected user-facing
		// behavior for deleting DBs.
		if _, found := foundDBFiles[path]; !found {
			if err := db.Close(); err != nil {
				return fmt.Errorf("closing deleted database: %w", err)
			}
			delete(a.dbs, path)
		}
	}
	newDBPaths := []string{}
	for _, dbPath := range dbPaths {
		if _, found := a.dbs[dbPath]; !found {
			newDBPaths = append(newDBPaths, dbPath)
		}
	}

	if len(newDBPaths) == 0 {
		// No new db-like files found that we haven't accounted for
		return nil
	}

	for _, dbPath := range newDBPaths {
		db, err := a.createDBFunc(dbPath)
		if err != nil {
			return fmt.Errorf("creating new db path=%q: %w", dbPath, err)
		}
		a.dbs[dbPath] = db
	}
	pl := pool.New().WithErrors()
	for _, db := range a.dbs {
		d := db // loop to goroutine
		pl.Go(func() error { return d.Sync(context.Background()) })
	}
	return pl.Wait()
}

func (a *application) newRequest(ctx context.Context) (err error) {
	if err := a.dbLimitWaiter.Wait(ctx); err != nil {
		return err
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.inFlightCounter++
	a.requestCount++
	if !a.running {
		a.startCount++
		a.running = true
		a.cmd, err = bunRun(a.dir, a.port, a.Env, nil)
		return err
	}
	return nil
}

func (a *application) endOfRequest() {
	a.mutex.Lock()
	a.inFlightCounter--
	if a.inFlightCounter == 0 {
		a.stopRequestChan <- struct{}{}
	}
	a.mutex.Unlock()
}

func bunRun(dir string, port int, env []string, logs io.Writer) (*execx.Cmd, error) {
	buffer := ring.NewBuffer(1024)
	var logWriter io.Writer = buffer
	if logs != nil {
		logWriter = io.MultiWriter(buffer, logs)
	}

	wrapperPath := filepath.Join(dir, "wrapper.ts")
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		if err := os.WriteFile(wrapperPath, wrapperTS, 0666); err != nil {
			return nil, err
		}
	}

	healthEndpoint := steadyutil.RandomString(10)
	cmd := execx.Command("bun", "run", "--no-install", "wrapper.ts")
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, []string{
		"STEADY_INDEX_LOCATION=/opt/app/index.ts",
		"STEADY_HEALTH_ENDPOINT=/" + healthEndpoint,
		fmt.Sprintf("PORT=%d", port)}...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stderr = logWriter
	cmd.Stdout = logWriter

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	count := 20
	for i := 0; i < count; i++ {
		// TODO: replace this all with a custom version of Bun so that we don't
		// impact user applications, or a wrapper script
		req, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", port, healthEndpoint))
		if err == nil {
			_ = req.Body.Close()
			break
		}

		if !cmd.Running() {
			maybeErrLogs := string(buffer.Bytes())
			err := fmt.Errorf("Exited with code %d", cmd.ExitCode())
			if len(maybeErrLogs) > 0 {
				err = fmt.Errorf("%s\n%s", err.Error(), maybeErrLogs)
			}
			return nil, err
		}

		if i == count-1 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			_ = cmd.Shutdown(ctx)
			cancel()
			return nil, fmt.Errorf("Process is running, but nothing is listening on the expected port")
		}
		exponent := time.Duration((i+1)*(i+1)) / 2
		time.Sleep(time.Millisecond * exponent)
	}
	return cmd, nil
}
