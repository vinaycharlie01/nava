package k3dmagex

import (
	"fmt"

	"github.com/nirantaraai/nava/pkg/k3d"
)

const defaultConfigPath = "k3d.yaml"

// Package-level runner, initialized by LoadConfig.
var defaultRunner *k3dx.K3dRunner

// LoadConfig loads k3d configuration from a YAML file.
func LoadConfig(filepath string) error {
	runner, err := k3dx.NewK3dRunnerFromYAML(filepath)
	if err != nil {
		return err
	}
	defaultRunner = runner
	return nil
}

// LoadDefaultConfig loads from the default "k3d.yaml" path.
func LoadDefaultConfig() error {
	return LoadConfig(defaultConfigPath)
}

func ensureLoaded() error {
	if defaultRunner == nil {
		return fmt.Errorf("k3d config not loaded; call LoadConfig() or LoadDefaultConfig() first")
	}
	return nil
}

// ---------- Cluster ----------

// ClusterCreate creates the k3d cluster.
func ClusterCreate() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.ClusterCreate()
}

// ClusterDelete deletes the k3d cluster.
func ClusterDelete() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.ClusterDelete()
}

// ClusterList lists k3d clusters.
func ClusterList() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.ClusterList()
}

// ---------- Bootstrap ----------

// Bootstrap applies infrastructure kustomization and StorageClass.
func Bootstrap() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.Bootstrap()
}

// ---------- Helm ----------

// HelmRepos adds and updates all configured Helm repositories.
func HelmRepos() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.HelmRepos()
}

// InstallRelease installs a single release by name.
func InstallRelease(name string) error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.InstallRelease(name)
}

// InstallAllReleases installs all releases in order.
func InstallAllReleases() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.InstallAllReleases()
}

// UninstallRelease uninstalls a release by name.
func UninstallRelease(name string) error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.UninstallRelease(name)
}

// ---------- ArgoCD ----------

// InstallArgoCD installs ArgoCD.
func InstallArgoCD() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.InstallArgoCD()
}

// CreateRepoSecret creates ArgoCD repo secret for private git access.
func CreateRepoSecret() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.CreateRepoSecret()
}

// CreateBitbucketRepoSecret creates ArgoCD repo secret for Bitbucket private git access.
func CreateBitbucketRepoSecret() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.CreateBitbucketRepoSecret()
}

// ApplyAppOfApps applies the ArgoCD app-of-apps manifest.
func ApplyAppOfApps() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.ApplyAppOfApps()
}

// PatchApplication patches an ArgoCD application.
func PatchApplication(name, patchType, patch string) error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.PatchApplication(name, patchType, patch)
}

// ---------- Composite ----------

// Up creates cluster, bootstraps, and installs all releases.
func Up() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.Up()
}

// Down tears down the cluster.
func Down() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.Down()
}

// GitopsBootstrap creates cluster, installs ArgoCD, applies app-of-apps.
func GitopsBootstrap() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.GitopsBootstrap()
}

// Status shows cluster state.
func Status() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.Status()
}

// ReleaseNames returns configured release names.
func ReleaseNames() ([]string, error) {
	if err := ensureLoaded(); err != nil {
		return nil, err
	}
	return defaultRunner.ReleaseNames(), nil
}

// ApplicationsToPrune returns configured application names to prune.
func ApplicationsToPrune() ([]string, error) {
	if err := ensureLoaded(); err != nil {
		return nil, err
	}
	return defaultRunner.ApplicationsToPrune(), nil
}
