//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	gomagex "github.com/nirantaraai/nava/mage/golang"
	helmmagex "github.com/nirantaraai/nava/mage/helm"
	komagex "github.com/nirantaraai/nava/mage/ko"
)

// init loads the YAML configs once before any target runs.
func init() {
	_ = gomagex.LoadConfig("go.yaml")
	_ = helmmagex.LoadConfig("helm.yaml")
	_ = komagex.LoadConfig("ko.yaml")
}

// Go namespace for Go development targets
type Go mg.Namespace

// Build compiles the Go packages
func (Go) Build() error { return gomagex.Build() }

// Test runs the test suite
func (Go) Test() error { return gomagex.Test() }

// Race runs tests with race detection
func (Go) Race() error { return gomagex.Race() }

// Coverage runs tests with coverage profiling
func (Go) Coverage() error { return gomagex.Coverage() }

// Bench runs benchmarks
func (Go) Bench() error { return gomagex.Bench() }

// Lint runs golangci-lint
func (Go) Lint() error { return gomagex.Lint() }

// Vet runs go vet
func (Go) Vet() error { return gomagex.Vet() }

// Govulncheck runs govulncheck
func (Go) Govulncheck() error { return gomagex.Govulncheck() }

// Setup downloads Go module dependencies
func (Go) Setup() error { return gomagex.Setup() }

// Helm namespace for Helm-related targets
type Helm mg.Namespace

// Install installs a Helm chart
func (Helm) Install() error { return helmmagex.Install() }

// Upgrade upgrades a Helm release
func (Helm) Upgrade() error { return helmmagex.Upgrade() }

// Uninstall uninstalls a Helm release
func (Helm) Uninstall() error { return helmmagex.Uninstall() }

// List lists all Helm releases
func (Helm) List() error { return helmmagex.List() }

// Lint lints a Helm chart
func (Helm) Lint() error { return helmmagex.Lint() }

// RepoUpdate updates Helm repositories
func (Helm) RepoUpdate() error { return helmmagex.RepoUpdate() }

// Ko namespace for Ko (container building) targets
type Ko mg.Namespace

// Build builds a container image with ko
func (Ko) Build() error { return komagex.Build() }

// Apply builds images and applies Kubernetes manifests
func (Ko) Apply() error { return komagex.Apply() }

// Delete deletes Kubernetes resources
func (Ko) Delete() error { return komagex.Delete() }

// Publish publishes a container image
func (Ko) Publish() error { return komagex.Publish() }
