package steady

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"golang.org/x/sync/errgroup"
)

var exampleServer = `
export default {
	port: 3000,
	fetch(request) {
		return new Response("Hello %s");
	},
};
`

type TestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment

	now time.Time
	w   *Workflow
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.w = NewWorkflow()

	testSuite := &testsuite.WorkflowTestSuite{}
	s.env = testSuite.NewTestWorkflowEnvironment()
	s.env.RegisterWorkflow(s.w.Workflow)
	s.now = time.Date(2022, time.December, 1, 0, 0, 0, 0, time.UTC)
	s.env.SetStartTime(s.now)
	s.env.RegisterActivity(s.w.RunWorkerActivity)

	s.env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		TaskQueue: workflowQueueName,
	})
}

func (s *TestSuite) TestSimpleRequest() {
	var meta Meta

	s.env.RegisterDelayedCallback(func() {
		meta = s.queryMeta()
	}, time.Second)

	s.env.RegisterDelayedCallback(func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", meta.Port))
		if err != nil {
			s.T().Fatal(err)
		}
		s.Equal(resp.StatusCode, http.StatusOK)
	}, time.Second*2)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("updates", Update{Stop: true})
	}, time.Second*3)

	s.env.ExecuteWorkflow(s.w.Workflow, exampleServer)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func testHTTPGet(url string) (body string, resp *http.Response, err error) {
	if resp, err = http.Get(url); err != nil {
		return "", resp, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if resp.Body != nil {
		_, _ = io.Copy(&buf, resp.Body)
	}
	return buf.String(), resp, err
}

func (s *TestSuite) TestNewApplication() {
	var meta Meta

	timestamp := time.Now().Format(time.RFC3339)

	s.env.RegisterDelayedCallback(func() {
		meta = s.queryMeta()
	}, time.Second)

	s.env.RegisterDelayedCallback(func() {
		body, resp, err := testHTTPGet(fmt.Sprintf("http://localhost:%d", meta.Port))
		if err != nil {
			s.T().Fatal(err)
		}
		s.Equal(resp.StatusCode, http.StatusOK)
	}, time.Second*2)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("updates", Update{Stop: true})
	}, time.Second*3)

	s.env.ExecuteWorkflow(s.w.Workflow, exampleServer)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *TestSuite) queryMeta() (meta Meta) {
	value, err := s.env.QueryWorkflow("metadata")
	if err != nil {
		s.T().Fatal(err)
	}
	if err := value.Get(&meta); err != nil {
		s.T().Error(err)
	}
	return meta
}

func TestConcurrentRequests(t *testing.T) {
	w := NewWorkflow()
	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })

	requestCount := 5
	for i := 0; i < requestCount; i++ {
		j := i
		eg.Go(func() error {
			resp, err := http.Get("http://localhost:8080")
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			io.Copy(&buf, resp.Body)
			assert.Contains(t, buf.String(), timestamp)
			if j == 4 {
				cancel()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	cancel()

	assert.Equal(t, requestCount, w.workerState.requestCount)
	assert.Equal(t, 1, w.workerState.startCount)
}

func TestNonOverlappingTests(t *testing.T) {
	w := NewWorkflow()
	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		t.Fatal(err)
	}

	makeRequest := func() {
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(t, buf.String(), timestamp)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })

	makeRequest()

	// Wait for shutdown
	// TODO: parameterize
	time.Sleep(time.Second)

	makeRequest()
	cancel()

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, w.workerState.requestCount)
	assert.Equal(t, 2, w.workerState.startCount)
}

func BenchmarkActivity(b *testing.B) {
	w := NewWorkflow()
	w.requestLogger = io.Discard

	timestamp := time.Now().Format(time.RFC3339)
	filename, err := w.writeApplicationScript(
		fmt.Sprintf(exampleServer, timestamp),
	)
	if err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return w.RunWorkerActivity(ctx, WorkerData{Filename: filename}) })
	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			b.Fatal(err)
		}
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		assert.Contains(b, buf.String(), timestamp)
	}
	cancel()
	if err := eg.Wait(); err != nil {
		b.Fatal(err)
	}
}
