package ca

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Config identifies the targets we're going to look for signing requests in
type Config struct {
	Projects []string `yaml:"Projects"`
	Folders  []string `yaml:"Folders"`
}

func loadConfigFile(file string) Config {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("Unable to read file: " + file)
	}
	var c Config
	if err = yaml.Unmarshal(yamlFile, &c); err != nil {
		log.Fatalln("Unable to parse yaml file: " + file)
	}
	return c
}
