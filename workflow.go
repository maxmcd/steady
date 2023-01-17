package steady

import (
	"io"
	"net"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	workflowQueueName = "STEADY_TODO"
)

type Option func(*Workflow)

func NewClient() (client.Client, error) {
	hostPort := os.Getenv("TEMPORAL_GRPC_ENDPOINT")
	if hostPort == "" {
		hostPort = client.DefaultHostPort
	}

	return client.NewClient(client.Options{
		HostPort: hostPort,
	})
}

func NewWorkflow(options ...Option) *Workflow {
	s := &Workflow{
		requestLogger: os.Stdout,
	}
	for _, option := range options {
		option(s)
	}
	return s
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
	wkr.RegisterActivity(w.RunWorkerActivity)

	// This will block until its interrupted
	return wkr.Run(worker.InterruptCh())
}

type Workflow struct {
	workerState   *WorkerState
	requestLogger io.Writer
	port          int
}

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

type Update struct {
	Application *string
	Stop        bool
}

func (w *Workflow) Workflow(ctx workflow.Context, application string) (err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// https://docs.temporal.io/docs/concepts/what-is-a-start-to-close-timeout
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		// https://docs.temporal.io/application-development/features/#activity-retry-simulator
		// TODO: Consider removing? We don't retry?
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: 2,
			InitialInterval:    1000,
			MaximumAttempts:    20,
		},
	})

	logger := workflow.GetLogger(ctx)
	filename, err := w.writeApplicationScript(application)
	if err != nil {
		return err
	}

	w.port, err = getFreePort()
	if err != nil {
		return err
	}

	if err := workflow.SetQueryHandler(ctx, "metadata", func() (Meta, error) {
		return Meta{
			Port: w.port,
		}, nil
	}); err != nil {
		return err
	}

	selector := workflow.NewSelector(ctx)
	updatesChannel := workflow.GetSignalChannel(ctx, "updates")

	workflow.Go(ctx, func(ctx workflow.Context) {
		if err := workflow.ExecuteActivity(ctx, w.RunWorkerActivity, WorkerData{
			Port:     w.port,
			Filename: filename,
		}).Get(ctx, nil); err != nil {
			logger.Error("Found unexpected activity error", "err", err)
		}
	})

	var update Update
	selector.AddReceive(updatesChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &update)
	})
	for {
		// continually listen for signals
		selector.Select(ctx)
		if update.Application != nil {
			return workflow.NewContinueAsNewError(ctx,
				w.Workflow,
				*update.Application,
			)
		}
		if update.Stop {
			return nil
		}
	}
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	_ = l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
