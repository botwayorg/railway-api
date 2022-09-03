package cmd

import (
	"context"
	"fmt"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/ui"
)

func (h *Handler) Unlink(ctx context.Context, req *entity.CommandRequest) error {
	projectCfg, _ := h.ctrl.GetProjectConfigs(ctx)

	project, err := h.ctrl.GetProject(ctx, projectCfg.Project)
	if err != nil {
		return err
	}

	err = h.cfg.RemoveProjectConfigs(projectCfg)
	if err != nil {
		return err
	}

	fmt.Printf("🎉 Disconnected from %s\n", ui.MagentaText(project.Name))
	return nil
}
