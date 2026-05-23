package util

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Token   string `yaml:"token"`
	BaseUrl string `yaml:"base_url"`
	Bucket  string `yaml:"bucket"`
}

func LoadConfig(path string) (Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	return config, err
}
