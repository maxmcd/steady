package daemon

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/benbjohnson/litestream"
)

type Application struct {
	name string
	// port is the port this application listens on
	port int
	// dir is the directory that contains all application data
	dir string

	DBs []string
	Env []string

	mutex           sync.Mutex
	inFlightCounter int

	cmd     *exec.Cmd
	running bool

	stopRequestChan chan struct{}

	requestCount int
	startCount   int

	cancel func()

	litestreamServer *litestream.Server
	createDBFunc     func(string) (*litestream.DB, error)
}

func (d *Daemon) newApplication(name string, dir string, port int) *Application {
	w := &Application{
		name:            name,
		dir:             dir,
		port:            port,
		stopRequestChan: make(chan struct{}),
		createDBFunc:    d.createDB(name),
	}
	return w
}

func (a *Application) Start() error {
	if err := a.checkForDBs(); err != nil {
		return err
	}

	go a.runLoop()
	return nil
}

func (a *Application) runLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	killTimer := time.NewTimer(math.MaxInt64)
	for {
		// TODO: stop loop
		select {
		case <-a.stopRequestChan:
			killTimer.Reset(time.Second)
		case <-killTimer.C:
			a.stopProcess()
		case <-ctx.Done():
			return
		}
	}
}

func (a *Application) stopProcess() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if !a.running {
		return
	}
	_ = a.cmd.Process.Kill()
	a.running = false
	if err := a.checkForDBs(); err != nil {
		fmt.Println("Warn: ", err)
	}
}

func (a *Application) shutdown() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cancel()
	if a.running {
		_ = a.cmd.Process.Kill()
		a.running = false
	}
	if a.litestreamServer != nil {
		return a.litestreamServer.Close()
	}
	return nil
}

func (a *Application) checkForDBs() error {
	dbs, err := filepath.Glob(filepath.Join(a.dir, "./*.sql"))
	if err != nil {
		return err
	}
	foundDBFiles := map[string]struct{}{}
	for _, db := range dbs {
		foundDBFiles[db] = struct{}{}
	}

	existingPaths := map[string]struct{}{}

	if a.litestreamServer != nil {
		for _, db := range a.litestreamServer.DBs() {
			path := db.Path()
			// If we didn't find a litestream db in our project, unwatch it.
			// TODO: Is this even possible?
			if _, found := foundDBFiles[path]; !found {
				if err := a.litestreamServer.Unwatch(path); err != nil {
					return err
				}
			} else {
				existingPaths[path] = struct{}{}
			}
		}
	}
	newDBs := []string{}
	for _, db := range dbs {
		if _, found := existingPaths[db]; !found {
			newDBs = append(newDBs, db)
		}
	}

	if len(newDBs) == 0 {
		// No new db-like files found that we haven't accounted for
		return nil
	}

	if a.litestreamServer == nil {
		a.litestreamServer = litestream.NewServer()
		if err := a.litestreamServer.Open(); err != nil {
			return err
		}
	}
	for _, db := range newDBs {
		if err := a.litestreamServer.Watch(db, a.createDBFunc); err != nil {
			return err
		}
	}
	for _, db := range a.litestreamServer.DBs() {
		if _, found := existingPaths[db.Path()]; found {
			// Don't init db's we already have
			continue
		}
		// Sync now so that we can catch errors
		if err := db.Sync(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func bunRun(dir string, port int, env []string) (*exec.Cmd, error) {
	cmd := exec.Command("bun", "index.ts")
	cmd.Dir = dir
	cmd.Env = append([]string{fmt.Sprintf("PORT=%d", port)}, env...)

	// TODO: log to file
	var buf bytes.Buffer
	cmd.Stderr = &buf
	cmd.Stdout = &buf
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	// Mutex to prevent race on access of process state both while ending the
	// process and also during out checks
	lock := sync.Mutex{}
	var processState *os.ProcessState

	go func() {
		// Ensure ProcessState is populated in the event of a failure
		_ = cmd.Wait()
		lock.Lock()
		processState = cmd.ProcessState
		lock.Unlock()
	}()
	count := 20
	for i := 0; i < count; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			_ = conn.Close()
			break
		}
		exited := false
		exitCode := 0
		lock.Lock()
		// TODO: clean up!
		exists := processState != nil
		if exists {
			exitCode = processState.ExitCode()
			exited = processState.Exited()
		}
		lock.Unlock()
		if exists && exited {
			if exitCode != 0 {
				if len(buf.Bytes()) == 0 {
					return nil, fmt.Errorf("Exited with code %d", exitCode)
				}
				return nil, fmt.Errorf(buf.String())
			}
			return nil, fmt.Errorf("Exited with code %d", exitCode)
		}

		if i == count-1 {
			_ = cmd.Process.Kill()
			return nil, fmt.Errorf("Process is running, but nothing is listening on the expected port")
		}
		exponent := time.Duration(i + 1*i + 1)
		time.Sleep(time.Millisecond * exponent)
	}
	return cmd, nil
}

func (a *Application) newRequest() (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.inFlightCounter++
	a.requestCount++
	if !a.running {
		a.startCount++
		a.running = true
		a.cmd, err = bunRun(a.dir, a.port, a.Env)
		return err
	}
	return nil
}

func (a *Application) endOfRequest() {
	a.mutex.Lock()
	a.inFlightCounter--
	if a.inFlightCounter == 0 {
		a.stopRequestChan <- struct{}{}
	}
	a.mutex.Unlock()
}
