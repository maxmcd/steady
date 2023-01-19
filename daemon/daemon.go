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
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

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

func (d *Daemon) addApplication(name, script string) error {
	port, err := getFreePort()
	if err != nil {
		return errors.Wrap(err, "error attempting to get free port")
	}

	dirName := d.applicationDirectory(name)
	if err := os.Mkdir(dirName, 0777); err != nil {
		return errors.Wrapf(err, "error creating directory %q", dirName)
	}
	fileName := filepath.Join(dirName, "index.ts")
	if err := os.WriteFile(fileName, []byte(script), 0666); err != nil {
		return errors.Wrapf(err, "error creating file %q", fileName)
	}
	app := newApplication(dirName, port)

	d.applicationsLock.Lock()
	d.applications[name] = app
	d.applicationsLock.Unlock()
	return nil
}

func (d *Daemon) validateApplicationScript(script []byte) (dir string, err error) {
	dir, err = os.MkdirTemp(d.dataDirectory, "")
	if err != nil {
		return "", err
	}
	fileName := filepath.Join(dir, "index.ts")
	if err := os.WriteFile(fileName, script, 0666); err != nil {
		return "", errors.Wrapf(err, "error creating file %q", fileName)
	}

	port, err := getFreePort()
	if err != nil {
		return "", err
	}

	cmd, err := bunRun(dir, port)
	if err != nil {
		return "", err
	}
	return dir, cmd.Process.Kill() // maybe ignore kill err?
}

func (d *Daemon) Start(ctx context.Context) {
	if d.eg != nil {
		panic("Daemon has already started")
	}
	d.eg, ctx = errgroup.WithContext(ctx)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	steadyGroup := r.Group("/steady")

	steadyGroup.POST("/application/:name", func(c *gin.Context) {
		script, err := io.ReadAll(c.Request.Body)
		if err != nil || len(script) == 0 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"msg": "request body must contain a valid application script"})
			return
		}

		if dir, err := d.validateApplicationScript(script); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"msg": err.Error()})
			return
		}
		name := c.Param("name")
		d.applicationsLock.Lock()
		_, found := d.applications[name]
		if !found {
			app = &Application{}
			d.applications[name] = app
		}
		d.applicationsLock.Unlock()

		if found {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"msg": "an application with this name is already present on this host"})
			return
		}
	})
	steadyGroup.DELETE("/application/:name", func(ctx *gin.Context) {

	})
	steadyGroup.POST("/application/:name/migrate", func(ctx *gin.Context) {

	})

	steadyGroup.Use(func(ctx *gin.Context) {
		if len(ctx.Errors.Errors()) > 0 {

		}
	})

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

func (w *Application) startProcess() error {
	// Mutex must be acquired when this is called
	w.startCount++
	w.running = true

	return nil
}

func (w *Application) newRequest() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.inFlightCounter++
	w.requestCount++
	if !w.running {
		if err := w.startProcess(); err != nil {
			return err
		}
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
