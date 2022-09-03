package cmd

import (
	"context"
	"fmt"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/ui"
)

func (h *Handler) Login(ctx context.Context, req *entity.CommandRequest) error {
	isBrowserless, err := req.Cmd.Flags().GetBool("browserless")
	if err != nil {
		return err
	}

	user, err := h.ctrl.Login(ctx, isBrowserless)
	if err != nil {
		return err
	}

	fmt.Printf("\n🎉 Logged in as %s (%s)\n", ui.Bold(user.Name), user.Email)

	return nil
}
