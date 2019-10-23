package main

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

//HGNConfig struct used to consume configuration details
type HGNConfig struct {
	CertFile    string `yaml:"certFile"`
	CertKeyFile string `yaml:"certKeyFile"`

	BotName  string `yaml:"botName"`
	MasterID string `yaml:"masterID"`

	Port string `yaml:"port"`
}

//DBConfig struct used to consume Configuration details
//regarding database access
type DBConfig struct {
	Driver string `yaml:"driver"`
	DBUser string `yaml:"user"`
	DBName string `yaml:"name"`
	DBPass string `yaml:"password"`
	DBOpts string `yaml:"options"`
}

//loadConfig specifcially loads configuration information
//for the bot as opposed to the database.
//I'm pretty sure there is a better way to use one function
//to load and distribute configs where they are needed, but
//I haven't quite figreud out the way to do so just yet.
func loadConfig(src string) (config HGNConfig) {
	configB, err := ioutil.ReadFile(src)
	checkError(err)

	yaml.Unmarshal(configB, &config)

	return
}

//loadDBConfig specifically loads configuration information
//for the database as oppposed to the bot/connections
func loadDBConfig(src string) (config DBConfig) {
	configB, err := ioutil.ReadFile(src)
	checkError(err)

	yaml.Unmarshal(configB, &config)

	return
}
