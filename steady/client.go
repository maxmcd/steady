package steady

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type Client struct {
	temporalClient client.Client
}

type User struct {
	Email string
}

func (c *Client) NewUser(ctx context.Context, user User) error {
	run, err := c.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}, new(UserWorkflow).Workflow)
	if err != nil {
		return err
	}
	fmt.Println(run.GetID())
	return nil
}
