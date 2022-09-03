package controller

import (
	"github.com/botwayorg/railway-api/configs"
	"github.com/botwayorg/railway-api/gateway"
	"github.com/botwayorg/railway-api/random"
	"github.com/google/go-github/github"
)

type Controller struct {
	gtwy       *gateway.Gateway
	cfg        *configs.Configs
	randomizer *random.Randomizer
	ghc        *github.Client
}

func New() *Controller {
	return &Controller{
		gtwy:       gateway.New(),
		cfg:        configs.New(),
		randomizer: random.New(),
		ghc:        github.NewClient(nil),
	}
}
