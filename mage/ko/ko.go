package komagex

import (
	kox "github.com/nirantaraai/nava/pkg/ko"
)

// Package-level runner for mage targets
var defaultRunner = kox.NewKoRunner()

// LoadConfig loads ko configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*kox.KoRunner, error) {
	return kox.NewKoRunnerFromYAML(filepath)
}

// Build builds a container image with ko (requires loaded config)
func Build() error {
	return defaultRunner.Build()
}

// Apply builds images and applies Kubernetes manifests (requires loaded config)
func Apply() error {
	return defaultRunner.Apply()
}

// Delete deletes Kubernetes resources (requires loaded config)
func Delete() error {
	return defaultRunner.Delete()
}

// Resolve resolves import paths to image references (requires loaded config)
func Resolve() error {
	return defaultRunner.Resolve()
}

// Publish publishes a container image (requires loaded config)
func Publish() error {
	return defaultRunner.Publish()
}
