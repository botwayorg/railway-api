package gateway

import (
	context "context"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/errors"
)

func (g *Gateway) GetWorkflowStatus(ctx context.Context, workflowID string) (entity.WorkflowStatus, error) {
	gqlReq, err := g.NewRequestWithAuth(`
		query($workflowId: String!) {
			getWorkflowStatus(workflowId: $workflowId) {
				status
			}
		}
	`)
	if err != nil {
		return "", err
	}

	gqlReq.Var("workflowId", workflowID)

	var resp struct {
		WorkflowStatus *entity.WorkflowStatusResponse `json:"getWorkflowStatus"`
	}
	if err := gqlReq.Run(ctx, &resp); err != nil {
		return "", errors.ProjectCreateFailed
	}
	return resp.WorkflowStatus.Status, nil
}
