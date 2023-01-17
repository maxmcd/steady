package steady

import "go.temporal.io/sdk/workflow"

type ApplicationWorkflow struct{}

func (s *ApplicationWorkflow) Workflow(ctx workflow.Context) (err error) {
	// Each host has a slice of hash ranges that it controls.

	return nil
}
