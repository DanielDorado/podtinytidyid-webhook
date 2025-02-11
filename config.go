package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server    Server
	Generator Generator
}

type Server struct {
	Port int
	TLS  TLSCert `yaml:"TLS"`
}

type TLSCert struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type Generator struct {
	IdentifierBits int `yaml:"identifierBits"`
}

var SingletonConfig *Config = nil

func NewConfigFromFile(file string) (*Config, error) {
	if SingletonConfig != nil {
		return SingletonConfig, nil
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return &Config{}, fmt.Errorf("Getting configuration: %e", err)
	}
	return NewConfig(raw)
}

func NewConfig(raw []byte) (*Config, error) {
	config := &Config{}

	err := yaml.Unmarshal(raw, config)
	if err != nil {
		return config, fmt.Errorf("Getting configuration: %e", err)
	}
	return config, nil
}
