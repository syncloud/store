package util

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Token               string `yaml:"token"`
	BaseUrl             string `yaml:"base_url"`
	Bucket              string `yaml:"bucket"`
	AwsAccessKeyId      string `yaml:"aws_access_key_id"`
	AwsSecretAccessKey  string `yaml:"aws_secret_access_key"`
	AwsS3Endpoint       string `yaml:"aws_s3_endpoint"`
	AwsRegion           string `yaml:"aws_region"`
	SigningKeyActive    string `yaml:"signing_key_active"`
	SigningKeyNewBase64 string `yaml:"signing_key_new_base64"`
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
