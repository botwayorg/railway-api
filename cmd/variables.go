package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/ui"
)

func (h *Handler) Variables(ctx context.Context, req *entity.CommandRequest) error {
	serviceName, err := req.Cmd.Flags().GetString("service")

	if err != nil {
		return err
	}

	envs, err := h.ctrl.GetEnvsForCurrentEnvironment(ctx, &serviceName)

	if err != nil {
		return err
	}

	environment, err := h.ctrl.GetCurrentEnvironment(ctx)

	if err != nil {
		return err
	}

	fmt.Print(ui.Heading(fmt.Sprintf("%s Environment Variables", environment.Name)))
	fmt.Print(ui.KeyValues(*envs, false))

	return nil
}

func (h *Handler) VariablesGet(ctx context.Context, req *entity.CommandRequest) error {
	serviceName, err := req.Cmd.Flags().GetString("service")

	if err != nil {
		return err
	}

	envs, err := h.ctrl.GetEnvsForCurrentEnvironment(ctx, &serviceName)

	if err != nil {
		return err
	}

	for _, key := range req.Args {
		fmt.Println(envs.Get(key))
	}

	return nil
}

func (h *Handler) VariablesSet(ctx context.Context, req *entity.CommandRequest) error {
	serviceName, err := req.Cmd.Flags().GetString("service")
	if err != nil {
		return err
	}

	skipRedeploy, err := req.Cmd.Flags().GetBool("skip-redeploy")
	if err != nil {
		// The flag is optional; default to false.
		skipRedeploy = false
	}

	replace, err := req.Cmd.Flags().GetBool("replace")
	if err != nil {
		// The flag is optional; default to false.
		replace = false
	}

	yes, err := req.Cmd.Flags().GetBool("yes")
	if err != nil {
		// The flag is optional; default to false.
		yes = false
	}

	if replace && !yes {
		fmt.Println(ui.Bold(ui.RedText(fmt.Sprintf("Warning! You are about to fully replace all your variables for the service '%s'.", serviceName)).String()))
		confirm, err := ui.PromptYesNo("Continue?")

		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}
	}

	variables := &entity.Envs{}
	updatedEnvNames := make([]string, 0)

	for _, kvPair := range req.Args {
		parts := strings.SplitN(kvPair, "=", 2)

		if len(parts) != 2 {
			return errors.New("invalid variables invocation. See --help")
		}

		key := parts[0]
		value := parts[1]

		variables.Set(key, value)
		updatedEnvNames = append(updatedEnvNames, key)
	}

	err = h.ctrl.UpdateEnvs(ctx, variables, &serviceName, replace)

	if err != nil {
		return err
	}

	environment, err := h.ctrl.GetCurrentEnvironment(ctx)
	if err != nil {
		return err
	}

	operation := "Updated"
	if replace {
		operation = "Replaced existing variables with"
	}

	fmt.Print(ui.Heading(fmt.Sprintf("%s %s for \"%s\"", operation, strings.Join(updatedEnvNames, ", "), environment.Name)))
	fmt.Print(ui.KeyValues(*variables, false))

	if !skipRedeploy {
		serviceID, err := h.ctrl.GetServiceIdByName(ctx, &serviceName)
		if err != nil {
			return err
		}

		err = h.redeployAfterVariablesChange(ctx, environment, serviceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) VariablesDelete(ctx context.Context, req *entity.CommandRequest) error {
	serviceName, err := req.Cmd.Flags().GetString("service")
	if err != nil {
		return err
	}

	skipRedeploy, err := req.Cmd.Flags().GetBool("skip-redeploy")
	if err != nil {
		// The flag is optional; default to false.
		skipRedeploy = false
	}

	err = h.ctrl.DeleteEnvs(ctx, req.Args, &serviceName)
	if err != nil {
		return err
	}

	environment, err := h.ctrl.GetCurrentEnvironment(ctx)
	if err != nil {
		return err
	}

	fmt.Print(ui.Heading(fmt.Sprintf("Deleted %s for \"%s\"", strings.Join(req.Args, ", "), environment.Name)))

	if !skipRedeploy {
		serviceID, err := h.ctrl.GetServiceIdByName(ctx, &serviceName)
		if err != nil {
			return err
		}

		err = h.redeployAfterVariablesChange(ctx, environment, serviceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) redeployAfterVariablesChange(ctx context.Context, environment *entity.Environment, serviceID *string) error {
	deployments, err := h.ctrl.GetDeployments(ctx)
	if err != nil {
		return err
	}

	// Don't redeploy if we don't yet have any deployments
	if len(deployments) == 0 {
		return nil
	}

	// Don't redeploy if the latest deploy for environment came from up
	latestDeploy := deployments[0]
	if latestDeploy.Meta == nil || latestDeploy.Meta.Repo == "" {
		fmt.Printf(ui.AlertInfo("Run %s to redeploy your project"), ui.MagentaText("railway up").Underline())
		return nil
	}

	ui.StartSpinner(&ui.SpinnerCfg{
		Message: fmt.Sprintf("Redeploying \"%s\" with new variables", environment.Name),
	})

	err = h.ctrl.DeployEnvironmentTriggers(ctx, serviceID)
	if err != nil {
		return err
	}

	ui.StopSpinner("Deploy triggered")

	deployment, err := h.ctrl.GetLatestDeploymentForEnvironment(ctx, latestDeploy.ProjectID, environment.Id)
	if err != nil {
		return err
	}

	fmt.Printf("☁️ Deploy Logs available at %s\n", ui.GrayText(h.ctrl.GetServiceDeploymentsURL(ctx, latestDeploy.ProjectID, *serviceID, deployment.ID)))

	return nil
}
