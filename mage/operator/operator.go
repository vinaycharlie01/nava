// Package operatormagex exposes package-level convenience functions for Mage
// targets that build and deploy Kubernetes operators.  It wraps pkg/operator.
package operatormagex

import (
	operatorx "github.com/nirantaraai/nava/pkg/operator"
)

// defaultRunner is the package-level runner used by all convenience functions.
var defaultRunner = operatorx.NewOperatorRunner()

// LoadConfig loads operator configuration from a YAML file into the default runner.
func LoadConfig(path string) error {
	return defaultRunner.LoadConfig(path)
}

// SetLocalBin overrides the directory where tool binaries are cached (default: ./bin).
func SetLocalBin(dir string) {
	defaultRunner.LocalBin = dir
}

// ─── Tool bootstrapping ───────────────────────────────────────────────────────

// EnsureControllerGen downloads controller-gen into bin/ if not already present.
func EnsureControllerGen(version string) error {
	return defaultRunner.EnsureControllerGen(version)
}

// EnsureKustomize downloads kustomize into bin/ if not already present.
func EnsureKustomize(version string) error {
	return defaultRunner.EnsureKustomize(version)
}

// EnsureEnvtest downloads setup-envtest into bin/ if not already present.
func EnsureEnvtest(version string) error {
	return defaultRunner.EnsureEnvtest(version)
}

// ─── Code/manifest generation ────────────────────────────────────────────────

// GenerateManifests runs controller-gen to produce CRD, RBAC, and webhook manifests.
func GenerateManifests() error {
	return defaultRunner.GenerateManifests()
}

// GenerateCode runs controller-gen to produce DeepCopy methods.
// When boilerplate is empty, no headerFile argument is passed (bare object generation).
func GenerateCode(boilerplate string) error {
	return defaultRunner.GenerateCode(boilerplate)
}

// GenerateAll runs DeepCopy generation, manifest generation, and CRD sync in one call.
func GenerateAll(boilerplate, helmCRDDir string) error {
	return defaultRunner.GenerateAll(boilerplate, helmCRDDir)
}

// GenerateFromConfig runs code generation using loaded YAML config values.
func GenerateFromConfig() error {
	return defaultRunner.GenerateFromConfig()
}

// GenerateAndSync generates manifests and copies CRDs to the Helm chart crds/ directory.
func GenerateAndSync(helmCRDDir string) error {
	return defaultRunner.GenerateAndSync(helmCRDDir)
}

// SyncCRDs copies CRD files from srcDir to the Helm chart destDir.
func SyncCRDs(srcDir, destDir string) error {
	return defaultRunner.SyncCRDs(srcDir, destDir)
}

// ─── Deployment ──────────────────────────────────────────────────────────────

// InstallCRDs installs CRDs into the cluster (kustomize build config/crd | kubectl apply).
func InstallCRDs() error {
	return defaultRunner.InstallCRDs()
}

// UninstallCRDs removes CRDs from the cluster.
func UninstallCRDs() error {
	return defaultRunner.UninstallCRDs()
}

// Deploy deploys the operator via kustomize build config/default | kubectl apply.
// Pass image="" to keep the existing kustomize image setting.
func Deploy(image string) error {
	return defaultRunner.Deploy(image)
}

// Undeploy removes the operator from the cluster.
func Undeploy() error {
	return defaultRunner.Undeploy()
}

// BuildInstaller generates dist/install.yaml (all CRDs + Deployment in one file).
func BuildInstaller(image string) error {
	return defaultRunner.BuildInstaller(image)
}

// ─── Clean ───────────────────────────────────────────────────────────────────

// Clean removes build artifacts. Uses config.Clean.Paths when loaded; otherwise
// removes the supplied paths.
func Clean(paths ...string) error {
	return defaultRunner.Clean(paths...)
}

// CleanFromConfig removes build artifacts using the loaded YAML config.
func CleanFromConfig() error {
	return defaultRunner.CleanFromConfig()
}

// ─── Testing ─────────────────────────────────────────────────────────────────

// EnvtestAssets downloads envtest binaries for the given Kubernetes version
// (e.g. "1.36") and returns the path to set as KUBEBUILDER_ASSETS.
func EnvtestAssets(k8sVersion string) (string, error) {
	return defaultRunner.EnvtestAssets(k8sVersion)
}
