package controller

import (
	"context"

	"github.com/botwayorg/railway-api/entity"
)

func (c *Controller) Down(ctx context.Context, req *entity.DownRequest) error {
	err := c.gtwy.Down(ctx, req)

	return err
}
