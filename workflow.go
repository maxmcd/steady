package steady

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	workflowQueueName = "STEADY_TODO"
)

type Option func(*Workflow)

func NewWorkflow(options ...Option) *Workflow {
	s := &Workflow{}
	for _, option := range options {
		option(s)
	}
	return s
}

func NewClient() (client.Client, error) {
	hostPort := os.Getenv("TEMPORAL_GRPC_ENDPOINT")
	if hostPort == "" {
		hostPort = client.DefaultHostPort
	}

	return client.NewClient(client.Options{
		HostPort: hostPort,
	})
}
func (w *Workflow) StartWorker() error {
	c, err := NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	wkr := worker.New(c, workflowQueueName, worker.Options{
		// Uncomment if you want to see logs of tasks when they are replayed,
		// helpful for replay errors and general debugging
		// EnableLoggingInReplay: true,
	})
	wkr.RegisterWorkflow(w.Workflow)
	wkr.RegisterActivity(w.RunWorker)

	// This will block until its interrupted
	return wkr.Run(worker.InterruptCh())
}

type Workflow struct{}

func (w *Workflow) writeApplicationScript(application string) (filename string, err error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(application); err != nil {
		return "", err
	}
	f.Close()

	return f.Name(), nil
}

func (w *Workflow) Workflow(ctx workflow.Context, application string) (err error) {
	logger := workflow.GetLogger(ctx)
	var inFlightRequests int

	filename, err := w.writeApplicationScript(application)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if err := workflow.SetQueryHandler(ctx, "request", func(Request) (Response, error) {
		var err error
		workerState.startWorkerIfDead(func() {
		})
		if err != nil {
			return Response{}, err
		}

		workerState.startWorkerIfDead(func() {
			err = cmd.Process.Signal(os.Interrupt)
		})

		// TODO:
		// - send request to running server if it is running
		// - start server if it is not running
		// - terminate server if this is the final request (you can use memory in the worker to track requests!)
		return Response{}, nil
	}); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, w.RunWorker, filename).Get(ctx, nil); err != nil {
		logger.Error("Found unexpected activity error", "err", err)
		return err
	}
	return nil
}

type WorkerState struct {
	mutex           sync.Mutex
	inFlightCounter int
}

func (w *WorkerState) startWorkerIfDead(cb func()) {
	w.mutex.Lock()
	if w.inFlightCounter == 0 {
		cb()
	}
	w.inFlightCounter++
	w.mutex.Unlock()
}
func (w *WorkerState) stopWorkerIfDone(cb func()) {
	w.mutex.Lock()
	if w.inFlightCounter == 1 {
		cb()
		w.inFlightCounter--
	}
	w.mutex.Unlock()
}

// RunWorker just loops indefinitely until the context is killed, reserving a
// spot on the server
func (w *Workflow) RunWorker(ctx context.Context, filename string) (err error) {
	workerState := &WorkerState{}
	var cmd *exec.Cmd
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workerState.startWorkerIfDead(func() {
			cmd = exec.Command("bun", filename)
			cmd.Env = []string{"PORT=9000"}
			err = cmd.Start()
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		workerState.stopWorkerIfDone(func() {})
	}))

	select {
	case <-ctx.Done():
		return nil
	}
}
