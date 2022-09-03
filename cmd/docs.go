package cmd

import (
	"context"

	"github.com/botwayorg/railway-api/constants"
	"github.com/botwayorg/railway-api/entity"
)

func (h *Handler) Docs(ctx context.Context, req *entity.CommandRequest) error {
	return h.ctrl.ConfirmBrowserOpen("Opening Railway Docs...", constants.RailwayDocsURL)
}
