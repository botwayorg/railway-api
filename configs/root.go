package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"

    "github.com/botwayorg/railway-api/errors"
	"github.com/botwayorg/railway-api/entity"
)

func (c *Configs) GetRootConfigs() (*entity.RootConfig, error) {
	var cfg entity.RootConfig

	b, err := ioutil.ReadFile(c.rootConfigs.configPath)

	if os.IsNotExist(err) {
        return nil, errors.RootConfigNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &cfg)

	return &cfg, err
}

func (c *Configs) SetRootConfig(cfg *entity.RootConfig) error {
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]entity.ProjectConfig)
	}

	return c.marshalConfig(c.rootConfigs, *cfg)
}
