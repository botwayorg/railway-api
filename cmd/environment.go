package cmd

import (
	"context"
	"fmt"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/ui"
	"github.com/manifoldco/promptui"
)

func (h *Handler) Environment(ctx context.Context, req *entity.CommandRequest) error {
	projectID, err := h.cfg.GetProject()

	if err != nil {
		return err
	}

	project, err := h.ctrl.GetProject(ctx, projectID)

	if err != nil {
		return err
	}

	var environment *entity.Environment

	if len(req.Args) > 0 {
		var name = req.Args[0]

		// Look for existing environment with name
		for _, projectEnvironment := range project.Environments {
			if name == projectEnvironment.Name {
				environment = projectEnvironment
			}
		}

		if environment != nil {
			fmt.Printf("%s Environment: %s\n", promptui.IconGood, ui.BlueText(environment.Name))
		} else {
			// Create new environment
			environment, err = h.ctrl.CreateEnvironment(ctx, &entity.CreateEnvironmentRequest{
				Name:      name,
				ProjectID: project.Id,
			})

			if err != nil {
				return err
			}

			fmt.Printf("Created Environment %s\nEnvironment: %s\n", promptui.IconGood, ui.BlueText(ui.Bold(name).String()))
		}
	} else {
		// Existing environment selector
		environment, err = ui.PromptEnvironments(project.Environments)

		if err != nil {
			return err
		}
	}

	err = h.cfg.SetEnvironment(environment.Id)

	if err != nil {
		return err
	}

	fmt.Printf("%s ProTip: You can view the active environment by running %s\n", promptui.IconInitial, ui.BlueText("railway status"))

	return err
}
