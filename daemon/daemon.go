package daemon

import (
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

func (d *Daemon) Start(ctx context.Context) {
	if d.eg != nil {
		panic("Daemon has already started")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/steady/application", func(ctx *gin.Context) {
		fmt.Println("create")
	})
	r.DELETE("/steady/application", func(ctx *gin.Context) {
		fmt.Println("delete")
	})

	r.Any(":name/*path", d.applicationHandler)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: r,
	}

	d.eg, ctx = errgroup.WithContext(ctx)
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
}

type Application struct {
	port int
	dir  string

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

func (w *Application) startProcess() error {
	// Mutex must be acquired when this is called
	w.startCount++
	w.cmd = exec.Command("bun", "index.ts")
	w.cmd.Dir = w.dir
	w.cmd.Env = []string{fmt.Sprintf("PORT=%d", w.port)}
	err := w.cmd.Start()
	if err != nil {
		return err
	}
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", w.port))
		if err != nil {
			time.Sleep(time.Millisecond)
			// TODO: Error if the server never comes up
			continue
		}
		conn.Close()
		break
	}
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
