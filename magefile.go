//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	helmmagex "github.com/nirantaraai/nava/mage/helm"
	komagex "github.com/nirantaraai/nava/mage/ko"
)

// init loads the YAML configs once before any target runs. Errors (e.g. a
// missing file) are deferred to the target, which reports "configuration not
// loaded" only for the section it actually needs.
func init() {
	_ = helmmagex.LoadConfig("helm.yaml")
	_ = komagex.LoadConfig("ko.yaml")
}

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
