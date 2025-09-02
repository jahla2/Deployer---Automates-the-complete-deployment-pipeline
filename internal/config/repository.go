package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"deployer/internal/domain"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) LoadConfig(configFile string) (*domain.Config, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", configFile)
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config domain.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.SSH.Port == 0 {
		config.SSH.Port = 22
	}

	return &config, nil
}

func (r *Repository) GetServiceNames(config *domain.Config) []string {
	names := make([]string, 0, len(config.Services))
	for name := range config.Services {
		names = append(names, name)
	}
	return names
}