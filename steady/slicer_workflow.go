package steady

import "go.temporal.io/sdk/workflow"

type SlicerWorkflow struct{}

func (s *SlicerWorkflow) Workflow(ctx workflow.Context) (err error) {
	// Each host has a slice of hash ranges that it controls.

	return nil
}
