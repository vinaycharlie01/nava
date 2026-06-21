package goreleasermagex

import (
	goreleaserx "github.com/nirantaraai/nava/pkg/goreleaser"
)

// Package-level runner for mage targets
var defaultRunner = goreleaserx.NewGoReleaserRunner()

// LoadConfig loads goreleaser configuration from a YAML file
func LoadConfig(path string) error {
	return defaultRunner.LoadConfig(path)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(path string) (*goreleaserx.GoReleaserRunner, error) {
	return goreleaserx.NewGoReleaserRunnerFromYAML(path)
}

// Release runs goreleaser release (requires loaded config)
func Release() error { return defaultRunner.ReleaseFromConfig() }

// Snapshot runs a local snapshot release without publishing (requires loaded config)
func Snapshot() error { return defaultRunner.SnapshotFromConfig() }

// Check validates the goreleaser configuration (requires loaded config)
func Check() error { return defaultRunner.CheckFromConfig() }
