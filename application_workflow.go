package steady

import "go.temporal.io/sdk/workflow"

type ApplicationWorkflow struct{}

func (s *ApplicationWorkflow) Workflow(ctx workflow.Context, name string, source string) (err error) {
	// name is {user}/{application_name}

	// Find host, send application to host
	return nil
}

func (s *ApplicationWorkflow) PlaceApplicationActivity(ctx workflow.Context, name string, source string) (err error) {
	return nil
}
