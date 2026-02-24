package config

import "sync"

type RootConfig struct {
	Mutex   sync.RWMutex
	Address string `yaml:"address"`
	Port    uint   `yaml:"port"`
	ApiKey  string `yaml:"apikey"`
}

func DefaultRootConfig() *RootConfig {
	return &RootConfig{
		Address: "0.0.0.0",
		Port:    9112,
		ApiKey:  "",
	}
}
