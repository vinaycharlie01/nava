package k3dx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nirantaraai/nava/pkg/exec"
	"github.com/nirantaraai/nava/pkg/helm"
	"github.com/nirantaraai/nava/pkg/k8s"

	"gopkg.in/yaml.v3"
)

// K3dRunner handles k3d cluster lifecycle, Helm installs, and kubectl operations.
type K3dRunner struct {
	executor execx.Executor
	helm     *helmx.HelmRunner
	k8s      *k8sx.K8sRunner
	config   *K3dConfig
}

// NewK3dRunner creates a new K3dRunner with default executors.
func NewK3dRunner() *K3dRunner {
	executor := execx.NewExec()
	return &K3dRunner{
		executor: executor,
		helm:     helmx.NewHelmRunnerWithExecutor(executor),
		k8s:      k8sx.NewK8sRunnerWithExecutor(executor),
	}
}

// NewK3dRunnerWithExecutor creates a K3dRunner with a custom executor.
func NewK3dRunnerWithExecutor(executor execx.Executor) *K3dRunner {
	return &K3dRunner{
		executor: executor,
		helm:     helmx.NewHelmRunnerWithExecutor(executor),
		k8s:      k8sx.NewK8sRunnerWithExecutor(executor),
	}
}

// NewK3dRunnerFromYAML creates a runner with configuration loaded from YAML.
func NewK3dRunnerFromYAML(filepath string) (*K3dRunner, error) {
	runner := NewK3dRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads k3d configuration from a YAML file.
func (k *K3dRunner) LoadConfig(filepath string) error {
	config, err := LoadK3dConfig(filepath)
	if err != nil {
		return err
	}
	k.config = config
	return nil
}

// Config returns the loaded configuration.
func (k *K3dRunner) Config() *K3dConfig {
	return k.config
}

// ---------- Configuration types ----------

// K3dConfig is the root YAML configuration for a k3d-based stack.
type K3dConfig struct {
	Cluster        ClusterConfig       `yaml:"cluster"`
	StorageClass   *StorageClassConfig `yaml:"storageClass,omitempty"`
	Infrastructure InfraConfig         `yaml:"infrastructure"`
	HelmRepos      map[string]string   `yaml:"helmRepos"`
	Releases       []ReleaseConfig     `yaml:"releases"`
	ArgoCD         *ArgoCDConfig       `yaml:"argocd,omitempty"`
}

// ClusterConfig holds k3d cluster creation settings.
type ClusterConfig struct {
	Name    string   `yaml:"name"`
	Servers int      `yaml:"servers"`
	Agents  int      `yaml:"agents"`
	Ports   []string `yaml:"ports,omitempty"`
	K3sArgs []string `yaml:"k3sArgs,omitempty"`
}

// StorageClassConfig defines an optional StorageClass to create after cluster bootstrap.
type StorageClassConfig struct {
	Name                 string `yaml:"name"`
	Provisioner          string `yaml:"provisioner"`
	VolumeBindingMode    string `yaml:"volumeBindingMode,omitempty"`
	ReclaimPolicy        string `yaml:"reclaimPolicy,omitempty"`
	AllowVolumeExpansion bool   `yaml:"allowVolumeExpansion,omitempty"`
}

// InfraConfig points at the kustomize base that creates namespaces and secrets.
type InfraConfig struct {
	KustomizePath string `yaml:"kustomizePath"`
}

// ReleaseConfig represents a single Helm release to install.
type ReleaseConfig struct {
	Name       string   `yaml:"name"`
	Chart      string   `yaml:"chart"`
	Namespace  string   `yaml:"namespace"`
	BaseValues string   `yaml:"baseValues,omitempty"`
	EnvValues  string   `yaml:"envValues,omitempty"`
	Wait       bool     `yaml:"wait,omitempty"`
	Timeout    string   `yaml:"timeout,omitempty"`
	Set        []string `yaml:"set,omitempty"`
}

// ArgoCDConfig describes ArgoCD-specific settings.
type ArgoCDConfig struct {
	Release             ReleaseConfig     `yaml:"release"`
	AppOfApps           string            `yaml:"appOfApps"`
	RepoSecret          *RepoSecretConfig `yaml:"repoSecret,omitempty"`
	ApplicationsToPrune []string          `yaml:"applicationsToPrune,omitempty"`
}

// RepoSecretConfig holds private repo credentials for ArgoCD.
type RepoSecretConfig struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	SSHKeyPath string `yaml:"sshKeyPath"`
}

// LoadK3dConfig reads and parses a k3d YAML config file.
func LoadK3dConfig(filepath string) (*K3dConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read k3d config: %w", err)
	}
	var cfg K3dConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse k3d config: %w", err)
	}
	return &cfg, nil
}

// ---------- Cluster lifecycle ----------

// ClusterCreate creates the k3d cluster defined in config.
func (k *K3dRunner) ClusterCreate() error {
	c := k.config.Cluster
	slog.Info("🚀 Creating k3d cluster", "name", c.Name, "servers", c.Servers, "agents", c.Agents)
	start := time.Now()

	args := []string{
		"cluster", "create", c.Name,
		"--servers", fmt.Sprintf("%d", c.Servers),
		"--agents", fmt.Sprintf("%d", c.Agents),
	}
	for _, p := range c.Ports {
		args = append(args, "-p", p)
	}
	for _, a := range c.K3sArgs {
		args = append(args, "--k3s-arg", a)
	}

	if err := k.executor.Run(context.Background(), "k3d", false, args...); err != nil {
		return fmt.Errorf("failed to create cluster %s: %w", c.Name, err)
	}

	slog.Info("✅ Cluster created", "name", c.Name, "duration", time.Since(start))
	return nil
}

// ClusterDelete deletes the k3d cluster.
func (k *K3dRunner) ClusterDelete() error {
	name := k.config.Cluster.Name
	slog.Info("🗑️  Deleting k3d cluster", "name", name)

	if err := k.executor.Run(context.Background(), "k3d", false, "cluster", "delete", name); err != nil {
		return fmt.Errorf("failed to delete cluster %s: %w", name, err)
	}

	slog.Info("✅ Cluster deleted", "name", name)
	return nil
}

// ClusterList lists k3d clusters.
func (k *K3dRunner) ClusterList() error {
	return k.executor.Run(context.Background(), "k3d", false, "cluster", "list")
}

// ---------- Bootstrap ----------

// Bootstrap applies infrastructure kustomization and optionally creates StorageClass.
func (k *K3dRunner) Bootstrap() error {
	slog.Info("🏗️  Bootstrapping infrastructure", "path", k.config.Infrastructure.KustomizePath)

	if err := k.k8s.Apply(k8sx.ApplyOptions{Kustomize: k.config.Infrastructure.KustomizePath}); err != nil {
		return fmt.Errorf("bootstrap kustomize failed: %w", err)
	}

	if k.config.StorageClass != nil {
		if err := k.createStorageClass(); err != nil {
			return err
		}
	}

	slog.Info("✅ Bootstrap complete")
	return nil
}

func (k *K3dRunner) createStorageClass() error {
	sc := k.config.StorageClass
	slog.Info("📦 Creating StorageClass", "name", sc.Name)

	vbm := sc.VolumeBindingMode
	if vbm == "" {
		vbm = "WaitForFirstConsumer"
	}
	rp := sc.ReclaimPolicy
	if rp == "" {
		rp = "Delete"
	}

	manifest := fmt.Sprintf(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: %s
provisioner: %s
volumeBindingMode: %s
reclaimPolicy: %s
allowVolumeExpansion: %v`, sc.Name, sc.Provisioner, vbm, rp, sc.AllowVolumeExpansion)

	return k.k8s.Apply(k8sx.ApplyOptions{Stdin: manifest})
}

// ---------- Helm repos ----------

// HelmRepos adds and updates all configured Helm repositories.
func (k *K3dRunner) HelmRepos() error {
	slog.Info("📚 Adding Helm repositories", "count", len(k.config.HelmRepos))

	for name, url := range k.config.HelmRepos {
		if err := k.helm.RepoAddNamed(name, url); err != nil {
			slog.Warn("Repo add failed (may already exist)", "name", name, "error", err)
		}
	}

	return k.helm.RepoUpdate()
}

// ---------- Helm installs ----------

// InstallRelease installs a single Helm release by name from the config.
func (k *K3dRunner) InstallRelease(name string) error {
	for _, r := range k.config.Releases {
		if r.Name == name {
			return k.helmInstall(r)
		}
	}
	return fmt.Errorf("release %q not found in config", name)
}

// InstallAllReleases installs all configured Helm releases in order.
func (k *K3dRunner) InstallAllReleases() error {
	for _, r := range k.config.Releases {
		if err := k.helmInstall(r); err != nil {
			return err
		}
	}
	return nil
}

// UninstallRelease uninstalls a Helm release by name.
func (k *K3dRunner) UninstallRelease(name string) error {
	for _, r := range k.config.Releases {
		if r.Name == name {
			return k.helm.UninstallRelease(r.Name, r.Namespace)
		}
	}
	return fmt.Errorf("release %q not found in config", name)
}

func (k *K3dRunner) helmInstall(r ReleaseConfig) error {
	slog.Info("📦 Installing release", "name", r.Name, "chart", r.Chart, "namespace", r.Namespace)
	start := time.Now()

	var values []string
	if r.BaseValues != "" {
		values = append(values, r.BaseValues)
	}
	if r.EnvValues != "" {
		values = append(values, r.EnvValues)
	}

	if err := k.helm.UpgradeWith(helmx.UpgradeOptions{
		ReleaseName:     r.Name,
		Chart:           r.Chart,
		Namespace:       r.Namespace,
		Values:          values,
		Set:             r.Set,
		Install:         true,
		CreateNamespace: true,
		Wait:            r.Wait,
		Timeout:         r.Timeout,
	}); err != nil {
		return fmt.Errorf("failed to install %s: %w", r.Name, err)
	}

	slog.Info("✅ Release installed", "name", r.Name, "duration", time.Since(start))
	return nil
}

// ---------- ArgoCD ----------

// InstallArgoCD installs ArgoCD using the config's argocd section.
func (k *K3dRunner) InstallArgoCD() error {
	if k.config.ArgoCD == nil {
		return fmt.Errorf("no argocd section in config")
	}
	return k.helmInstall(k.config.ArgoCD.Release)
}

// CreateRepoSecret creates a Kubernetes secret so ArgoCD can access a private git repo via SSH.
func (k *K3dRunner) CreateRepoSecret() error {
	if k.config.ArgoCD == nil || k.config.ArgoCD.RepoSecret == nil {
		slog.Info("⏭️  No argocd.repoSecret configured, skipping")
		return nil
	}
	rs := k.config.ArgoCD.RepoSecret

	// Expand ~ in SSH key path
	keyPath := rs.SSHKeyPath
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home dir: %w", err)
		}
		keyPath = home + keyPath[1:]
	}

	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("SSH key not found at %s: %w", keyPath, err)
	}

	slog.Info("🔑 Creating ArgoCD repo secret", "name", rs.Name, "url", rs.URL)

	// Delete existing secret if present (ignore error)
	_ = k.k8s.Delete(k8sx.DeleteOptions{Namespace: "argocd"}, "secret", rs.Name)

	if err := k.k8s.CreateSecret(k8sx.CreateSecretOptions{
		Name:      rs.Name,
		Namespace: "argocd",
		Type:      "generic",
		FromLiteral: []string{
			"type=git",
			"url=" + rs.URL,
		},
		FromFile: []string{
			"sshPrivateKey=" + keyPath,
		},
	}); err != nil {
		return fmt.Errorf("failed to create repo secret: %w", err)
	}

	if err := k.k8s.Label("secret", rs.Name, "argocd", map[string]string{
		"argocd.argoproj.io/secret-type": "repository",
	}, true); err != nil {
		return fmt.Errorf("failed to label repo secret: %w", err)
	}

	slog.Info("✅ Repo secret created", "name", rs.Name)
	return nil
}

// CreateBitbucketRepoSecret creates ArgoCD repo secret for Bitbucket using the configured SSH key.
// It reuses the same behavior as CreateRepoSecret to keep credentials handling in one place.
func (k *K3dRunner) CreateBitbucketRepoSecret() error {
	return k.CreateRepoSecret()
}

// ApplyAppOfApps applies the ArgoCD app-of-apps manifest.
func (k *K3dRunner) ApplyAppOfApps() error {
	if k.config.ArgoCD == nil || k.config.ArgoCD.AppOfApps == "" {
		return fmt.Errorf("no argocd.appOfApps path in config")
	}
	slog.Info("🔄 Applying ArgoCD app-of-apps", "path", k.config.ArgoCD.AppOfApps)
	return k.k8s.Apply(k8sx.ApplyOptions{
		Namespace: "argocd",
		Filenames: []string{k.config.ArgoCD.AppOfApps},
	})
}

// PatchApplication patches an ArgoCD application by name.
func (k *K3dRunner) PatchApplication(name, patchType, patch string) error {
	slog.Info("🩹 Patching ArgoCD application", "name", name, "patchType", patchType)
	return k.k8s.Patch("application", name, "argocd", patchType, patch)
}

// ---------- Composite workflows ----------

// Up creates cluster, adds repos, bootstraps, and installs all releases.
func (k *K3dRunner) Up() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"ClusterCreate", k.ClusterCreate},
		{"HelmRepos", k.HelmRepos},
		{"Bootstrap", k.Bootstrap},
		{"InstallAllReleases", k.InstallAllReleases},
	}
	return k.runSteps(steps)
}

// Down tears down the cluster.
func (k *K3dRunner) Down() error {
	return k.ClusterDelete()
}

// GitopsBootstrap creates cluster, bootstraps, installs ArgoCD, then applies app-of-apps.
func (k *K3dRunner) GitopsBootstrap() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"ClusterCreate", k.ClusterCreate},
		{"HelmRepos", k.HelmRepos},
		{"Bootstrap", k.Bootstrap},
		{"InstallArgoCD", k.InstallArgoCD},
		{"CreateRepoSecret", k.CreateRepoSecret},
		{"ApplyAppOfApps", k.ApplyAppOfApps},
	}
	return k.runSteps(steps)
}

// ---------- Status ----------

// Status runs kubectl commands to show cluster state.
func (k *K3dRunner) Status() error {
	_ = k.k8s.Get("nodes", "", "")
	_ = k.k8s.Get("pods", "", "argocd")
	_ = k.k8s.Get("pods", "", "ingress-nginx")
	_ = k.k8s.Get("pods", "", "monitoring")
	_ = k.k8s.Get("pvc", "", "monitoring")
	_ = k.k8s.Get("ingress", "", "", "-A")
	_ = k.k8s.Get("applications", "", "argocd")
	return nil
}

// ReleaseNames returns the list of configured release names.
func (k *K3dRunner) ReleaseNames() []string {
	names := make([]string, len(k.config.Releases))
	for i, r := range k.config.Releases {
		names[i] = r.Name
	}
	return names
}

// ApplicationsToPrune returns the list of configured application names to prune.
func (k *K3dRunner) ApplicationsToPrune() []string {
	if k.config.ArgoCD != nil {
		return k.config.ArgoCD.ApplicationsToPrune
	}
	return nil
}

// ---------- helpers ----------

func (k *K3dRunner) runSteps(steps []struct {
	name string
	fn   func() error
}) error {
	for _, s := range steps {
		slog.Info("--- " + s.name + " ---")
		if err := s.fn(); err != nil {
			return fmt.Errorf("%s failed: %w", s.name, err)
		}
	}
	return nil
}
