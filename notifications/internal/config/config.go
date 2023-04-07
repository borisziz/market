package config

import (
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
)

type ConfigStruct struct {
	Kafka struct {
		GroupName string   `yaml:"group_name"`
		Brokers   []string `yaml:"brokers"`
		Topics    []string `yaml:"topics"`
		Strategy  string   `yaml:"strategy"`
	} `yaml:"kafka"`
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
