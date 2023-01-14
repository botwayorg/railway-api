package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/errors"
	"github.com/botwayorg/railway-api/ui"
)

func (h *Handler) Unlink(ctx context.Context, _ *entity.CommandRequest) error {
	projectCfg, err := h.ctrl.GetProjectConfigs(ctx)
	if err == errors.ProjectConfigNotFound {
		fmt.Print(ui.AlertWarning("No project is currently linked"))
		os.Exit(1)
	} else if err != nil {
		return err
	}

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
