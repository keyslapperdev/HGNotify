package main

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type HGNConfig struct {
	CertFile    string `yaml:"certFile"`
	CertKeyFile string `yaml:"certKeyFile"`

	BotName  string `yaml:"botName"`
	MasterID string `yaml:"masterID"`

	Port string `yaml:"port"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DBUser string `yaml:"user"`
	DBName string `yaml:"name"`
	DBPass string `yaml:"password"`
	DBOpts string `yaml:"options"`
}

func loadConfig(src string) (config HGNConfig) {
	configB, err := ioutil.ReadFile(src)
	checkError(err)

	yaml.Unmarshal(configB, &config)

	return
}

func loadDBConfig(src string) (config DBConfig) {
	configB, err := ioutil.ReadFile(src)
	checkError(err)

	yaml.Unmarshal(configB, &config)

	return
}
