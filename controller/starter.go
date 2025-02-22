package controller

import (
	"context"

	"github.com/botwayorg/railway-api/entity"
)

// GetStarters returns all available starters
func (c *Controller) GetStarters(ctx context.Context) ([]*entity.Starter, error) {
	return c.gtwy.GetStarters(ctx)
}
