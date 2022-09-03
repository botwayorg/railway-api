package controller

import (
	"context"
	"github.com/botwayorg/railway-api/entity"
)

// GetWorkflowStatus fetches the status of a workflow based on request, error otherwise
func (c *Controller) GetWorkflowStatus(ctx context.Context, workflowID string) (entity.WorkflowStatus, error) {
	return c.gtwy.GetWorkflowStatus(ctx, workflowID)
}
