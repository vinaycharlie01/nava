package rundevx

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RunDevConfig represents the root configuration from rundev.yaml
type RunDevConfig struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

// ServiceConfig represents a single service configuration
type ServiceConfig struct {
	Language string `yaml:"language"`
	Filepath string `yaml:"filepath"`
}

// LoadRunDevConfig loads the rundev.yaml configuration
func LoadRunDevConfig(filepath string) (*RunDevConfig, error) {
	var config RunDevConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rundev.yaml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rundev.yaml: %w", err)
	}

	return &config, nil
}

// GetServiceConfig returns the configuration for a specific service
func GetServiceConfig(rundevPath, serviceName string) (*ServiceConfig, error) {
	config, err := LoadRunDevConfig(rundevPath)
	if err != nil {
		return nil, err
	}

	serviceConfig, ok := config.Services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found in rundev.yaml", serviceName)
	}

	return &serviceConfig, nil
}

// GetServiceFilepath returns the config filepath for a service
func GetServiceFilepath(rundevPath, serviceName string) (string, error) {
	serviceConfig, err := GetServiceConfig(rundevPath, serviceName)
	if err != nil {
		return "", err
	}

	return serviceConfig.Filepath, nil
}
