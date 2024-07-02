package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hosts    []string
	Endpoint string
}

func (c *Config) Init() {
	data, err := os.ReadFile("./conf/conf.yml")
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(data, c)
}
