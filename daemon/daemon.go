package daemon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
	"github.com/gin-gonic/gin"
	"github.com/maxmcd/steady/daemon/api"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

//go:generate bash -c "oapi-codegen --package=api --generate=types,client,gin,spec ./openapi3.yaml > api/api.gen.go"

type Daemon struct {
	dataDirectory string
	port          int
	client        http.Client

	s3Config *S3Config

	applicationsLock sync.RWMutex
	applications     map[string]*Application

	eg *errgroup.Group
}

type DaemonOption func(*Daemon)

func NewDaemon(dataDirectory string, port int, opts ...DaemonOption) *Daemon {
	d := &Daemon{
		dataDirectory: dataDirectory,
		port:          port,
		applications:  map[string]*Application{},
		client:        http.Client{},
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Path            string
	Endpoint        string
	SkipVerify      bool
	ForcePathStyle  bool
}

func DaemonOptionWithS3(cfg S3Config) func(*Daemon) { return func(d *Daemon) { d.s3Config = &cfg } }

func (d *Daemon) applicationDirectory(name string) string {
	return filepath.Join(d.dataDirectory, name)
}

func (d *Daemon) Wait() error { return d.eg.Wait() }

func (d *Daemon) s3Client() *awss3.S3 {
	if d.s3Config == nil {
		panic("no s3 config")
	}
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			d.s3Config.AccessKeyID,
			d.s3Config.SecretAccessKey,
			""),
		Endpoint:         aws.String(d.s3Config.Endpoint),
		Region:           aws.String("us-west-2"), // TODO
		DisableSSL:       aws.Bool(true),          // TODO
		S3ForcePathStyle: aws.Bool(d.s3Config.ForcePathStyle),
	}
	newSession := session.New(s3Config)

	return awss3.New(newSession)
}

func (d *Daemon) newReplica(db *litestream.DB, name string) *litestream.Replica {
	client := s3.NewReplicaClient()
	client.AccessKeyID = d.s3Config.AccessKeyID
	client.SecretAccessKey = d.s3Config.SecretAccessKey
	client.Bucket = d.s3Config.Bucket
	client.Path = filepath.Join(d.s3Config.Path, name)
	client.Endpoint = d.s3Config.Endpoint
	client.SkipVerify = d.s3Config.SkipVerify
	client.ForcePathStyle = d.s3Config.ForcePathStyle

	return litestream.NewReplica(db, name, client)
}

func (d *Daemon) createDB(name string) func(path string) (_ *litestream.DB, err error) {
	return func(path string) (_ *litestream.DB, err error) {
		if d.s3Config == nil {
			fmt.Println("WARN: skipping creating a new db because there is no s3 config")
			return nil, nil
		}
		db := litestream.NewDB(path)
		r := d.newReplica(db, filepath.Join(name, filepath.Base(path)))
		db.Replicas = append(db.Replicas, r)
		return db, nil
	}
}

func (d *Daemon) applicationHandler(c *gin.Context) {
	rw, r := c.Writer, c.Request

	name := c.Params.ByName("name")

	d.applicationsLock.RLock()
	app, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if !found {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	}

	// Remove name from path and route to correct port
	appURL := r.URL
	appURL.Path = c.Params.ByName("path")
	appURL.Host = fmt.Sprintf("localhost:%d", app.port)
	appURL.Scheme = "http"

	if err := app.newRequest(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer app.endOfRequest()

	{
		r.URL = appURL
		//http: Request.RequestURI can't be set in client requests.
		//http://golang.org/src/pkg/net/http/client.go

		r.RequestURI = ""
		// TODO: Set to expected path for application to see
		// req.Header.Set("Host", "TODO")

		resp, err := d.client.Do(r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = r.Body.Close()

		for k, v := range resp.Header {
			rw.Header()[k] = v
		}
		rw.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(rw, resp.Body)
		_ = resp.Body.Close()
	}

}

func (d *Daemon) validateAndAddApplication(name string, script []byte) (*Application, error) {
	// If two requests are validated simultaneously we won't catch them here,
	// but we'll error when adding them to the map later
	d.applicationsLock.RLock()
	_, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if found {
		return nil, fmt.Errorf("an application with this name is already present on this host")
	}

	tmpDir, err := os.MkdirTemp(d.dataDirectory, "")
	if err != nil {
		return nil, err
	}
	fileName := filepath.Join(tmpDir, "index.ts")
	if err := os.WriteFile(fileName, script, 0666); err != nil {
		return nil, errors.Wrapf(err, "error creating file %q", fileName)
	}

	port, err := getFreePort()
	if err != nil {
		return nil, err
	}

	cmd, err := bunRun(tmpDir, port)
	if err != nil {
		return nil, err
	}
	_ = cmd.Process.Kill()

	var dbs []string
	if d.s3Config != nil {
		s3Client := d.s3Client()
		output, err := s3Client.ListObjects(&awss3.ListObjectsInput{
			Bucket:    aws.String("litestream"), //TODO
			Prefix:    aws.String(name + "/"),
			Delimiter: aws.String("/"),
		})
		if err != nil {
			return nil, err
		}

		for _, dir := range output.CommonPrefixes {
			dbName := strings.TrimSuffix(
				strings.TrimPrefix(
					*dir.Prefix,
					name+"/"),
				"/",
			)
			dbs = append(dbs, dbName)
		}
	}

	for _, db := range dbs {
		dbPath := filepath.Join(tmpDir, db)

		_ = os.Remove(dbPath)
		// TODO: ensure we validate applications in an isolated environment
		// instead of just cleaning up after them
		//
		// TODO: also we just expect applications to run with an empty db? hmmm,
		// think about full lifecycle?

		replica := d.newReplica(nil, filepath.Join(name, db))
		generation, err := litestream.FindLatestGeneration(context.TODO(), replica.Client())
		if err != nil {
			panic(err)
		}
		targetIndex, err := litestream.FindMaxIndexByGeneration(context.TODO(), replica.Client(), generation)
		if err != nil {
			panic(err)
		}

		snapshotIndex, err := litestream.FindSnapshotForIndex(context.TODO(), replica.Client(), generation, targetIndex)
		if err != nil {
			panic(err)
		}
		if err := litestream.Restore(context.TODO(),
			replica.Client(),
			dbPath, generation,
			snapshotIndex, targetIndex,
			litestream.NewRestoreOptions()); err != nil {
			panic(err)
		}

	}

	app := d.newApplication(name, tmpDir, port, dbs)
	d.applicationsLock.Lock()
	if _, found = d.applications[name]; found {
		d.applicationsLock.Unlock()
		return nil, fmt.Errorf("an application with this name is already present on this host")
	}
	d.applications[name] = app
	d.applicationsLock.Unlock()

	return app, nil
}

func (d *Daemon) Start(ctx context.Context) {
	if d.eg != nil {
		panic("Daemon has already started")
	}
	d.eg, ctx = errgroup.WithContext(ctx)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api.RegisterHandlersWithOptions(r, server{daemon: d}, api.GinServerOptions{})
	r.Any(":name/*path", d.applicationHandler)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: r,
	}

	d.eg.Go(func() error {
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	})
	d.eg.Go(func() error {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			return err
		}
		d.applicationsLock.Lock()
		defer d.applicationsLock.Unlock()

		for _, app := range d.applications {
			// TODO: will need to be parallel if we ever care about shutdown
			// speed with live applications
			app.stopProcess()
		}

		return nil
	})

	// TODO: wait until server is live to exit?
}

type Application struct {
	name string
	// port is the port this application listens on
	port int
	// dir is the directory that contains all application data
	dir string

	mutex           sync.Mutex
	inFlightCounter int

	cmd     *exec.Cmd
	running bool

	C               chan struct{}
	stopRequestChan chan struct{}

	requestCount int
	startCount   int

	dbs              []string
	litestreamServer *litestream.Server
	createDBFunc     func(string) (*litestream.DB, error)
}

func (d *Daemon) newApplication(name string, dir string, port int, dbs []string) *Application {
	w := &Application{
		name:            name,
		dir:             dir,
		port:            port,
		stopRequestChan: make(chan struct{}),
		createDBFunc:    d.createDB(name),
		dbs:             dbs,
	}

	if err := w.setupDBs(); err != nil {
		panic(err)
	}

	go w.runLoop()
	return w
}

func (w *Application) runLoop() {
	killTimer := time.NewTimer(math.MaxInt64)
	for {
		// TODO: stop loop
		select {
		case <-w.stopRequestChan:
			killTimer.Reset(time.Second)
		case <-killTimer.C:
			w.stopProcess()
		}
	}
}

func (w *Application) stopProcess() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.running {
		return
	}
	_ = w.cmd.Process.Kill()
	w.running = false
	if err := w.checkForDBs(); err != nil {
		fmt.Println("Warn: ", err)
	}
}

func (w *Application) shutdown() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.running {
		return nil
	}
	_ = w.cmd.Process.Kill()
	w.running = false
	return w.litestreamServer.Close()
}

func (w *Application) setupDBs() error {
	if w.dbs == nil {
		return nil
	}
	w.litestreamServer = litestream.NewServer()
	if err := w.litestreamServer.Open(); err != nil {
		return err
	}
	for _, db := range w.dbs {
		dbPath := filepath.Join(w.dir, db)
		if err := w.litestreamServer.Watch(dbPath, w.createDBFunc); err != nil {
			return err
		}
	}
	for _, db := range w.litestreamServer.DBs() {
		if err := db.Sync(context.Background()); err != nil {
			return err
		}
	}
	return nil
}
func (w *Application) checkForDBs() error {
	if w.litestreamServer != nil {
		return nil
	}

	dbs, err := filepath.Glob(filepath.Join(w.dir, "./*.sql"))
	if err != nil {
		return err
	}

	if len(dbs) == 0 {
		return nil
	}

	w.litestreamServer = litestream.NewServer()
	if err := w.litestreamServer.Open(); err != nil {
		return err
	}
	for _, db := range dbs {
		if err := w.litestreamServer.Watch(db, w.createDBFunc); err != nil {
			return err
		}
	}
	for _, db := range w.litestreamServer.DBs() {
		if err := db.Sync(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func bunRun(dir string, port int) (*exec.Cmd, error) {
	cmd := exec.Command("bun", "index.ts")
	cmd.Dir = dir
	cmd.Env = []string{fmt.Sprintf("PORT=%d", port)}

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

func (w *Application) newRequest() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.inFlightCounter++
	w.requestCount++
	if !w.running {
		w.startCount++
		w.running = true
		w.cmd, err = bunRun(w.dir, w.port)
		return err
	}
	return nil
}

func (w *Application) endOfRequest() {
	w.mutex.Lock()
	w.inFlightCounter--
	if w.inFlightCounter == 0 {
		w.stopRequestChan <- struct{}{}
	}
	w.mutex.Unlock()
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
