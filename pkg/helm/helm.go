package helmx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// HelmRunner handles Helm command execution with dependency injection
type HelmRunner struct {
	executor execx.Executor
	config   *HelmConfig
}

// NewHelmRunner creates a new HelmRunner with the default executor
func NewHelmRunner() *HelmRunner {
	return &HelmRunner{
		executor: execx.NewExec(),
	}
}

// NewHelmRunnerWithExecutor creates a new HelmRunner with a custom executor
func NewHelmRunnerWithExecutor(executor execx.Executor) *HelmRunner {
	return &HelmRunner{
		executor: executor,
	}
}

// NewHelmRunnerFromYAML creates a new HelmRunner with configuration loaded from YAML
func NewHelmRunnerFromYAML(filepath string) (*HelmRunner, error) {
	config, err := LoadHelmConfig(filepath)
	if err != nil {
		return nil, err
	}
	return &HelmRunner{
		executor: execx.NewExec(),
		config:   config,
	}, nil
}

// LoadConfig loads Helm configuration from a YAML file
func (h *HelmRunner) LoadConfig(filepath string) error {
	config, err := LoadHelmConfig(filepath)
	if err != nil {
		return err
	}
	h.config = config
	return nil
}

// HelmConfig contains all Helm operation configurations
type HelmConfig struct {
	Install    *InstallOptions    `yaml:"install,omitempty"`
	Upgrade    *UpgradeOptions    `yaml:"upgrade,omitempty"`
	Uninstall  *UninstallOptions  `yaml:"uninstall,omitempty"`
	List       *ListOptions       `yaml:"list,omitempty"`
	Status     *StatusOptions     `yaml:"status,omitempty"`
	Template   *TemplateOptions   `yaml:"template,omitempty"`
	Lint       *LintOptions       `yaml:"lint,omitempty"`
	Package    *PackageOptions    `yaml:"package,omitempty"`
	RepoAdd    *RepoAddOptions    `yaml:"repoAdd,omitempty"`
	RepoUpdate *RepoUpdateOptions `yaml:"repoUpdate,omitempty"`
}

// LoadHelmConfig loads Helm configuration from a YAML file
func LoadHelmConfig(filepath string) (*HelmConfig, error) {
	var config HelmConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// InstallOptions contains options for helm install
type InstallOptions struct {
	ReleaseName     string   `yaml:"releaseName"`
	Chart           string   `yaml:"chart"`
	Namespace       string   `yaml:"namespace"`
	Values          []string `yaml:"values,omitempty"`          // --values or -f flags
	Set             []string `yaml:"set,omitempty"`             // --set flags
	CreateNamespace bool     `yaml:"createNamespace,omitempty"` // --create-namespace
	Wait            bool     `yaml:"wait,omitempty"`
	Timeout         string   `yaml:"timeout,omitempty"`
}

// Install installs a Helm chart using loaded configuration
func (h *HelmRunner) Install() error {
	if h.config == nil || h.config.Install == nil {
		return fmt.Errorf("install configuration not loaded")
	}
	return h.installChart(*h.config.Install)
}

func (h *HelmRunner) installChart(opts InstallOptions) error {
	if opts.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}
	if opts.Chart == "" {
		return fmt.Errorf("chart is required")
	}

	slog.Info("📦 Installing Helm chart...",
		"release", opts.ReleaseName,
		"chart", opts.Chart,
		"namespace", opts.Namespace,
	)

	start := time.Now()

	args := []string{"install", opts.ReleaseName, opts.Chart}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	if opts.CreateNamespace {
		args = append(args, "--create-namespace")
	}

	for _, valuesFile := range opts.Values {
		args = append(args, "--values", valuesFile)
	}

	for _, setValue := range opts.Set {
		args = append(args, "--set", setValue)
	}

	if opts.Wait {
		args = append(args, "--wait")
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout", opts.Timeout)
	}

	if err := h.executor.Run(context.Background(), "helm", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Helm chart installed", "duration", time.Since(start))
	return nil
}

// UpgradeOptions contains options for helm upgrade
type UpgradeOptions struct {
	ReleaseName     string   `yaml:"releaseName"`
	Chart           string   `yaml:"chart"`
	Namespace       string   `yaml:"namespace"`
	Values          []string `yaml:"values,omitempty"`
	Set             []string `yaml:"set,omitempty"`
	Install         bool     `yaml:"install,omitempty"` // --install flag
	CreateNamespace bool     `yaml:"createNamespace,omitempty"`
	Wait            bool     `yaml:"wait,omitempty"`
	Timeout         string   `yaml:"timeout,omitempty"`
}

// Upgrade upgrades a Helm release using loaded configuration
func (h *HelmRunner) Upgrade() error {
	if h.config == nil || h.config.Upgrade == nil {
		return fmt.Errorf("upgrade configuration not loaded")
	}
	return h.UpgradeWith(*h.config.Upgrade)
}

// UpgradeWith upgrades a Helm release with the supplied options. It is exported
// for programmatic callers that build options dynamically (e.g. the k3d runner).
func (h *HelmRunner) UpgradeWith(opts UpgradeOptions) error {
	if opts.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}
	if opts.Chart == "" {
		return fmt.Errorf("chart is required")
	}

	slog.Info("🔄 Upgrading Helm release...",
		"release", opts.ReleaseName,
		"chart", opts.Chart,
		"namespace", opts.Namespace,
	)

	start := time.Now()

	args := []string{"upgrade", opts.ReleaseName, opts.Chart}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	if opts.Install {
		args = append(args, "--install")
	}

	if opts.CreateNamespace {
		args = append(args, "--create-namespace")
	}

	for _, valuesFile := range opts.Values {
		args = append(args, "--values", valuesFile)
	}

	for _, setValue := range opts.Set {
		args = append(args, "--set", setValue)
	}

	if opts.Wait {
		args = append(args, "--wait")
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout", opts.Timeout)
	}

	if err := h.executor.Run(context.Background(), "helm", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Helm release upgraded", "duration", time.Since(start))
	return nil
}

// UninstallOptions contains options for helm uninstall
type UninstallOptions struct {
	ReleaseName string   `yaml:"releaseName"`
	Namespace   string   `yaml:"namespace"`
	Args        []string `yaml:"args,omitempty"`
}

// Uninstall uninstalls a Helm release using loaded configuration
func (h *HelmRunner) Uninstall() error {
	if h.config == nil || h.config.Uninstall == nil {
		return fmt.Errorf("uninstall configuration not loaded")
	}
	opts := h.config.Uninstall
	return h.UninstallRelease(opts.ReleaseName, opts.Namespace, opts.Args...)
}

// UninstallRelease uninstalls a Helm release by name. It is exported for
// programmatic callers (e.g. the k3d runner).
func (h *HelmRunner) UninstallRelease(releaseName, namespace string, args ...string) error {
	if releaseName == "" {
		return fmt.Errorf("release name is required")
	}

	slog.Info("🗑️  Uninstalling Helm release...",
		"release", releaseName,
		"namespace", namespace,
	)

	start := time.Now()

	cmdArgs := []string{"uninstall", releaseName}

	if namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", namespace)
	}

	cmdArgs = append(cmdArgs, args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm release uninstalled", "duration", time.Since(start))
	return nil
}

// ListOptions contains options for helm list
type ListOptions struct {
	Namespace string   `yaml:"namespace,omitempty"`
	Args      []string `yaml:"args,omitempty"`
}

// List lists Helm releases using loaded configuration
func (h *HelmRunner) List() error {
	var opts ListOptions
	if h.config != nil && h.config.List != nil {
		opts = *h.config.List
	}

	slog.Info("📋 Listing Helm releases...", "namespace", opts.Namespace)

	start := time.Now()

	cmdArgs := []string{"list"}

	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	} else {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	}

	cmdArgs = append(cmdArgs, opts.Args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm releases listed", "duration", time.Since(start))
	return nil
}

// StatusOptions contains options for helm status
type StatusOptions struct {
	ReleaseName string   `yaml:"releaseName"`
	Namespace   string   `yaml:"namespace"`
	Args        []string `yaml:"args,omitempty"`
}

// Status shows the status of a Helm release using loaded configuration
func (h *HelmRunner) Status() error {
	if h.config == nil || h.config.Status == nil {
		return fmt.Errorf("status configuration not loaded")
	}
	opts := h.config.Status
	if opts.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}

	slog.Info("📊 Getting Helm release status...",
		"release", opts.ReleaseName,
		"namespace", opts.Namespace,
	)

	start := time.Now()

	cmdArgs := []string{"status", opts.ReleaseName}

	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}

	cmdArgs = append(cmdArgs, opts.Args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm release status retrieved", "duration", time.Since(start))
	return nil
}

// TemplateOptions contains options for helm template
type TemplateOptions struct {
	ReleaseName string   `yaml:"releaseName"`
	Chart       string   `yaml:"chart"`
	Args        []string `yaml:"args,omitempty"`
}

// Template renders chart templates locally using loaded configuration
func (h *HelmRunner) Template() error {
	if h.config == nil || h.config.Template == nil {
		return fmt.Errorf("template configuration not loaded")
	}
	opts := h.config.Template
	if opts.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}
	if opts.Chart == "" {
		return fmt.Errorf("chart is required")
	}

	slog.Info("📝 Rendering Helm templates...",
		"release", opts.ReleaseName,
		"chart", opts.Chart,
	)

	start := time.Now()

	cmdArgs := []string{"template", opts.ReleaseName, opts.Chart}
	cmdArgs = append(cmdArgs, opts.Args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm templates rendered", "duration", time.Since(start))
	return nil
}

// LintOptions contains options for helm lint
type LintOptions struct {
	Chart string   `yaml:"chart"`
	Args  []string `yaml:"args,omitempty"`
}

// Lint runs helm lint on a chart using loaded configuration
func (h *HelmRunner) Lint() error {
	if h.config == nil || h.config.Lint == nil {
		return fmt.Errorf("lint configuration not loaded")
	}
	opts := h.config.Lint
	if opts.Chart == "" {
		return fmt.Errorf("chart path is required")
	}

	slog.Info("🔍 Linting Helm chart...", "chart", opts.Chart)

	start := time.Now()

	cmdArgs := []string{"lint", opts.Chart}
	cmdArgs = append(cmdArgs, opts.Args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm chart linted", "duration", time.Since(start))
	return nil
}

// PackageOptions contains options for helm package
type PackageOptions struct {
	Chart string   `yaml:"chart"`
	Args  []string `yaml:"args,omitempty"`
}

// Package packages a chart directory into a chart archive using loaded configuration
func (h *HelmRunner) Package() error {
	if h.config == nil || h.config.Package == nil {
		return fmt.Errorf("package configuration not loaded")
	}
	opts := h.config.Package
	if opts.Chart == "" {
		return fmt.Errorf("chart path is required")
	}

	slog.Info("📦 Packaging Helm chart...", "chart", opts.Chart)

	start := time.Now()

	cmdArgs := []string{"package", opts.Chart}
	cmdArgs = append(cmdArgs, opts.Args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm chart packaged", "duration", time.Since(start))
	return nil
}

// RepoAddOptions contains options for helm repo add
type RepoAddOptions struct {
	Name string   `yaml:"name"`
	URL  string   `yaml:"url"`
	Args []string `yaml:"args,omitempty"`
}

// RepoAdd adds a chart repository using loaded configuration
func (h *HelmRunner) RepoAdd() error {
	if h.config == nil || h.config.RepoAdd == nil {
		return fmt.Errorf("repoAdd configuration not loaded")
	}
	opts := h.config.RepoAdd
	return h.RepoAddNamed(opts.Name, opts.URL, opts.Args...)
}

// RepoAddNamed adds a chart repository by name and URL. It is exported for
// programmatic callers (e.g. the k3d runner).
func (h *HelmRunner) RepoAddNamed(name, url string, args ...string) error {
	if name == "" {
		return fmt.Errorf("repository name is required")
	}
	if url == "" {
		return fmt.Errorf("repository URL is required")
	}

	slog.Info("➕ Adding Helm repository...", "name", name, "url", url)

	start := time.Now()

	cmdArgs := []string{"repo", "add", name, url}
	cmdArgs = append(cmdArgs, args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm repository added", "duration", time.Since(start))
	return nil
}

// RepoUpdateOptions contains options for helm repo update
type RepoUpdateOptions struct {
	Args []string `yaml:"args,omitempty"`
}

// RepoUpdate updates chart repositories using loaded configuration
func (h *HelmRunner) RepoUpdate() error {
	var args []string
	if h.config != nil && h.config.RepoUpdate != nil {
		args = h.config.RepoUpdate.Args
	}

	slog.Info("🔄 Updating Helm repositories...")

	start := time.Now()

	cmdArgs := []string{"repo", "update"}
	cmdArgs = append(cmdArgs, args...)

	if err := h.executor.Run(context.Background(), "helm", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Helm repositories updated", "duration", time.Since(start))
	return nil
}

// Made with Bob
