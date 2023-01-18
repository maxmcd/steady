package daemon

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
	"golang.org/x/sync/errgroup"
)

type Daemon struct {
	dataDirectory string
	applications  map[string]Application
}

func NewDaemon(dataDirectory string) *Daemon {
	return &Daemon{
		dataDirectory: dataDirectory,
		applications:  map[string]Application{},
	}
}

func (d *Daemon) applicationDirectory(name string) string {
	return filepath.Join(d.dataDirectory, name)
}

func (d *Daemon) Run(port int) {

	client := &http.Client{}

	srv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if err := w.workerState.newRequest(); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			{
				req, err := http.NewRequest(r.Method, "http://localhost:3000/"+r.URL.RawPath, r.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}

				resp, err := client.Do(req)
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

			w.workerState.endOfRequest()
		}),
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	})
	eg.Go(func() error {
		count := 0
		heartbeatTicker := time.NewTicker(time.Millisecond * 500)
		defer heartbeatTicker.Stop()
		for {
			select {
			case <-heartbeatTicker.C:
				count++
				activity.RecordHeartbeat(ctx, count)
			case <-w.workerState.C:
			case <-ctx.Done():
				if err := srv.Shutdown(context.Background()); err != nil {
					return err
				}
				w.workerState.stopProcess()
				fmt.Println("CONTEXT DONE")
				return nil
			}
		}
	})
	return eg.Wait()
}

type Application struct {
	port int
	name string

	mutex           sync.Mutex
	inFlightCounter int

	filename string

	cmd     *exec.Cmd
	running bool

	C               chan struct{}
	stopRequestChan chan struct{}

	requestCount int
	startCount   int
}

func newAppLifecycle(filename string) *Application {
	w := &Application{
		filename:        filename,
		C:               make(chan struct{}),
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
		w.C <- struct{}{}
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
	w.cmd = exec.Command("bun", w.filename)
	w.cmd.Env = []string{"PORT=9000"}
	err := w.cmd.Start()
	if err != nil {
		return err
	}
	for {
		conn, err := net.Dial("tcp", "localhost:3000")
		if err != nil {
			time.Sleep(time.Millisecond)
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
