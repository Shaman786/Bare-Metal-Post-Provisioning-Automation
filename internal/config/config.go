package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type IPPool struct {
	Name     string   `yaml:"name"`
	CIDR     string   `yaml:"cidr"`
	Gateway  string   `yaml:"gateway"`
	DNS      []string `yaml:"dns"`
	Reserved []string `yaml:"reserved"`
}

type IPPoolConfig struct {
	Pools []IPPool `yaml:"pools"`
}

func Load(path string) IPPoolConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var cfg IPPoolConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
