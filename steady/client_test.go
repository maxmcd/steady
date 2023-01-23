package steady

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type RobodialTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
	now time.Time
}

func TestRobodialTestSuite(t *testing.T) {
	suite.Run(t, new(RobodialTestSuite))
}
func (s *RobodialTestSuite) SetupTest() {
	testSuite := &testsuite.WorkflowTestSuite{}
	s.env = testSuite.NewTestWorkflowEnvironment()
	s.now = time.Date(2022, time.December, 1, 0, 0, 0, 0, time.UTC)
	s.env.SetStartTime(s.now)
}

func (s *RobodialTestSuite) TestCreateUser() { s.env.RegisterWorkflow(new(UserWorkflow).Workflow) }
