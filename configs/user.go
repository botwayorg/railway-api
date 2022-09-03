package configs

import (
	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/errors"
)

func (c *Configs) GetUserConfigs() (*entity.UserConfig, error) {
	var rootCfg *entity.RootConfig
	rootCfg, err := c.GetRootConfigs()
	if err != nil {
		return nil, errors.UserConfigNotFound
	}

	if rootCfg.User.Token == "" {
		return nil, errors.UserConfigNotFound
	}

	return &rootCfg.User, nil
}

func (c *Configs) SetUserConfigs(cfg *entity.UserConfig) error {
	var rootCfg *entity.RootConfig
	rootCfg, err := c.GetRootConfigs()
	if err != nil {
		rootCfg = &entity.RootConfig{}
	}

	rootCfg.User = *cfg

	return c.SetRootConfig(rootCfg)
}
