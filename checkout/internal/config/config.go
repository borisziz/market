package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ConfigStruct struct {
	Token string `yaml:"token"`
	Ports struct {
		Http string `yaml:"http"`
		Grpc string `yaml:"grpc"`
	} `yaml:"ports"`
	DBConnectURL string `yaml:"db_connect_url"`
	Services     struct {
		Loms     string `yaml:"loms"`
		Products string `yaml:"products"`
	} `yaml:"services"`
	WorkerPool struct {
		Workers uint16 `yaml:"workers"`
		Retries uint8  `yaml:"retries"`
	} `yaml:"worker_pool"`
}

var ConfigData ConfigStruct

func Init() error {
	rawYAML, err := os.ReadFile("config.yml")
	if err != nil {
		return errors.WithMessage(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &ConfigData)
	if err != nil {
		return errors.WithMessage(err, "parsing yaml")
	}

	return nil
}
