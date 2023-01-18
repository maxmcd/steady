package steady

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
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
