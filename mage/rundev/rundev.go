package rundevmagex

import (
	"fmt"

	"github.com/nirantaraai/nava/mage/golang"
	"github.com/nirantaraai/nava/mage/nodejs"
	"github.com/nirantaraai/nava/mage/python"
	"github.com/nirantaraai/nava/pkg/rundev"
)

const defaultRunDevPath = "rundev.yaml"

// LoadConfig loads configuration for a service using rundev.yaml registry
// It automatically detects the language and loads the appropriate config
func LoadConfig(serviceName string) error {
	serviceConfig, err := rundevx.GetServiceConfig(defaultRunDevPath, serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service config: %w", err)
	}

	switch serviceConfig.Language {
	case "python":
		if err := pythonmagex.LoadConfig(serviceConfig.Filepath); err != nil {
			return fmt.Errorf("failed to load Python config: %w", err)
		}
	case "go":
		if err := golangmagex.LoadConfig(serviceConfig.Filepath); err != nil {
			return fmt.Errorf("failed to load Go config: %w", err)
		}
	case "nodejs":
		if err := nodejsmagex.LoadConfig(serviceConfig.Filepath); err != nil {
			return fmt.Errorf("failed to load Node.js config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported language: %s", serviceConfig.Language)
	}

	return nil
}
