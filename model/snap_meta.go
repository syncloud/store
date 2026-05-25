package model

import "gopkg.in/yaml.v2"

type SnapMeta struct {
	Name        string `yaml:"name"`
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

func ParseSnapMeta(b []byte) (SnapMeta, error) {
	var m SnapMeta
	err := yaml.Unmarshal(b, &m)
	return m, err
}
