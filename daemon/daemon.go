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
	"sync"
	"time"

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

	applicationsLock sync.RWMutex
	applications     map[string]*Application

	eg *errgroup.Group
}

func NewDaemon(dataDirectory string, port int) *Daemon {
	return &Daemon{
		dataDirectory: dataDirectory,
		port:          port,
		applications:  map[string]*Application{},
		client:        http.Client{},
	}
}

func (d *Daemon) applicationDirectory(name string) string {
	return filepath.Join(d.dataDirectory, name)
}

func (d *Daemon) Wait() error { return d.eg.Wait() }

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
		req, err := http.NewRequest(r.Method, appURL.String(), r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO: Set to expected path for application to see
		// req.Header.Set("Host", "TODO")

		resp, err := d.client.Do(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		for k, v := range resp.Header {
			rw.Header()[k] = v
		}
		rw.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(rw, resp.Body)
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

	app := newApplication(tmpDir, port)
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
}

func newApplication(dir string, port int) *Application {
	w := &Application{
		dir:             dir,
		port:            port,
		stopRequestChan: make(chan struct{}),
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
	go func() { _ = cmd.Wait() }()
	count := 15
	for i := 0; i < count; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			_ = conn.Close()
			break
		}

		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			code := cmd.ProcessState.ExitCode()
			if cmd.ProcessState.ExitCode() != 0 {
				if len(buf.Bytes()) == 0 {
					return nil, fmt.Errorf("Exited with code %d", code)
				}
				return nil, fmt.Errorf(buf.String())
			}
			return nil, fmt.Errorf("Exited with code %d", code)
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
