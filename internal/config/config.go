package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Files []string `yaml:"files"`
	Rules []Rule   `yaml:"rules"`
}

type Rule struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

func LoadConfig(filename string) (*Config, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
