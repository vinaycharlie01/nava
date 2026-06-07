package helmmagex

import (
	helmx "github.com/nirantaraai/nava/pkg/helm"
)

// Package-level runner for mage targets
var defaultRunner = helmx.NewHelmRunner()

// LoadConfig loads Helm configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*helmx.HelmRunner, error) {
	return helmx.NewHelmRunnerFromYAML(filepath)
}

// Install installs a Helm chart (requires loaded config)
func Install() error {
	return defaultRunner.Install()
}

// Upgrade upgrades a Helm release (requires loaded config)
func Upgrade() error {
	return defaultRunner.Upgrade()
}

// Uninstall uninstalls a Helm release (requires loaded config)
func Uninstall() error {
	return defaultRunner.Uninstall()
}

// List lists Helm releases (requires loaded config)
func List() error {
	return defaultRunner.List()
}

// Status shows the status of a Helm release (requires loaded config)
func Status() error {
	return defaultRunner.Status()
}

// Template renders chart templates locally (requires loaded config)
func Template() error {
	return defaultRunner.Template()
}

// Lint runs helm lint on a chart (requires loaded config)
func Lint() error {
	return defaultRunner.Lint()
}

// Package packages a chart directory into a chart archive (requires loaded config)
func Package() error {
	return defaultRunner.Package()
}

// RepoAdd adds a chart repository (requires loaded config)
func RepoAdd() error {
	return defaultRunner.RepoAdd()
}

// RepoUpdate updates chart repositories (requires loaded config)
func RepoUpdate() error {
	return defaultRunner.RepoUpdate()
}
