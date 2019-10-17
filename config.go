package main

import (
    "io/ioutil"

    "github.com/go-yaml/yaml"
)

type HGNConfig struct {
    CertFile    string `yaml:"certFile"`
    CertKeyFile string `yaml:"certKeyFile"`

    Port string `yaml:"port"`

    BotName  string `yaml:"botName"`
    MasterID string `yaml:"MasterID"`
}

func loadConfig(src string) (config HGNConfig){
    configB, err := ioutil.ReadFile(src)
    checkError(err)

    yaml.Unmarshal(configB, &config)

    return
}
