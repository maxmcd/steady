package steady

import "go.temporal.io/sdk/workflow"

func userWorkflowID(user User) string {
	return "email-workflow-" + user.Email
}

type UserData struct {
	User         User
	Applications map[string]bool // bool just for serializability
}

type UserWorkflow struct{}

func (s *UserWorkflow) Workflow(ctx workflow.Context, UserData UserData) (err error) {

	return nil
}
