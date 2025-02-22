package cmd

import (
	"github.com/botwayorg/railway-api/configs"
	"github.com/botwayorg/railway-api/controller"
)

type Handler struct {
	ctrl *controller.Controller
	cfg  *configs.Configs
}

func New() *Handler {
	return &Handler{
		ctrl: controller.New(),
		cfg:  configs.New(),
	}
}
