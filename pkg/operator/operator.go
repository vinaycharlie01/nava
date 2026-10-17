// Package operatorx provides typed runners for Kubernetes operator build tooling:
// controller-gen (CRD/RBAC/webhook manifest generation + DeepCopy code generation),
// kustomize (manifest assembly), and setup-envtest (integration-test binary download).
//
// Every target in this package delegates to the named CLI tool via pkg/exec so
// output streams through to the terminal and the executor can be swapped in tests.
package operatorx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// OperatorConfig holds YAML-driven configuration for operator targets.
type OperatorConfig struct {
	LocalBin string          `yaml:"localBin,omitempty"`
	Generate *GenerateConfig `yaml:"generate,omitempty"`
	Clean    *CleanConfig    `yaml:"clean,omitempty"`
}

// GenerateConfig holds options for code/manifest generation.
type GenerateConfig struct {
	Boilerplate string `yaml:"boilerplate,omitempty"`
	HelmCRDDir  string `yaml:"helmCRDDir,omitempty"`
	CRDSrcDir   string `yaml:"crdSrcDir,omitempty"`
	APIPaths    string `yaml:"apiPaths,omitempty"`
}

// CleanConfig holds paths to clean up.
type CleanConfig struct {
	Paths []string `yaml:"paths,omitempty"`
}

// OperatorRunner wraps controller-gen, kustomize, and setup-envtest.
type OperatorRunner struct {
	executor execx.Executor
	// LocalBin is the directory where tool binaries are downloaded (default: ./bin).
	LocalBin string
	config   *OperatorConfig
}

// NewOperatorRunner creates a new OperatorRunner with the default executor and ./bin as LocalBin.
func NewOperatorRunner() *OperatorRunner {
	return &OperatorRunner{
		executor: execx.NewExec(),
		LocalBin: "bin",
	}
}

// NewOperatorRunnerWithExecutor creates a new OperatorRunner with a custom executor.
func NewOperatorRunnerWithExecutor(executor execx.Executor, localBin string) *OperatorRunner {
	return &OperatorRunner{executor: executor, LocalBin: localBin}
}

// LoadConfig loads operator configuration from a YAML file.
func (o *OperatorRunner) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read operator config: %w", err)
	}
	var cfg OperatorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal operator config: %w", err)
	}
	o.config = &cfg
	if cfg.LocalBin != "" {
		o.LocalBin = cfg.LocalBin
	}
	return nil
}

// LoadOperatorConfig loads operator configuration from a YAML file.
func LoadOperatorConfig(path string) (*OperatorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read operator config: %w", err)
	}
	var cfg OperatorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal operator config: %w", err)
	}
	return &cfg, nil
}

// controllerGen returns the path to the controller-gen binary.
// It checks LocalBin first, then falls back to PATH lookup.
func (o *OperatorRunner) controllerGen() string {
	localPath := filepath.Join(o.LocalBin, "controller-gen")
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	if p, err := exec.LookPath("controller-gen"); err == nil {
		return p
	}
	return localPath
}

// kustomize returns the path to the kustomize binary.
// It checks LocalBin first, then falls back to PATH lookup.
func (o *OperatorRunner) kustomize() string {
	localPath := filepath.Join(o.LocalBin, "kustomize")
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	if p, err := exec.LookPath("kustomize"); err == nil {
		return p
	}
	return localPath
}

// envtest returns the path to the setup-envtest binary.
// It checks LocalBin first, then falls back to PATH lookup.
func (o *OperatorRunner) envtest() string {
	localPath := filepath.Join(o.LocalBin, "setup-envtest")
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	if p, err := exec.LookPath("setup-envtest"); err == nil {
		return p
	}
	return localPath
}

// ─────────────────────────────────────────────────────────────────────────────
// Tool bootstrapping
// ─────────────────────────────────────────────────────────────────────────────

// EnsureControllerGen downloads controller-gen into LocalBin if not already present.
func (o *OperatorRunner) EnsureControllerGen(version string) error {
	bin := o.controllerGen()
	if _, err := os.Stat(bin); err == nil {
		slog.Info("controller-gen already present", "path", bin)
		return nil
	}
	slog.Info("⬇️  Downloading controller-gen", "version", version)
	if err := os.MkdirAll(o.LocalBin, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", o.LocalBin, err)
	}
	return o.executor.Run(context.Background(), "go", false,
		"install", fmt.Sprintf("sigs.k8s.io/controller-tools/cmd/controller-gen@%s", version))
}

// EnsureKustomize downloads kustomize into LocalBin if not already present.
func (o *OperatorRunner) EnsureKustomize(version string) error {
	bin := o.kustomize()
	if _, err := os.Stat(bin); err == nil {
		slog.Info("kustomize already present", "path", bin)
		return nil
	}
	slog.Info("⬇️  Downloading kustomize", "version", version)
	if err := os.MkdirAll(o.LocalBin, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", o.LocalBin, err)
	}
	return o.executor.Run(context.Background(), "go", false,
		"install", fmt.Sprintf("sigs.k8s.io/kustomize/kustomize/v5@%s", version))
}

// EnsureEnvtest downloads setup-envtest into LocalBin if not already present.
func (o *OperatorRunner) EnsureEnvtest(version string) error {
	bin := o.envtest()
	if _, err := os.Stat(bin); err == nil {
		slog.Info("setup-envtest already present", "path", bin)
		return nil
	}
	slog.Info("⬇️  Downloading setup-envtest", "version", version)
	if err := os.MkdirAll(o.LocalBin, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", o.LocalBin, err)
	}
	return o.executor.Run(context.Background(), "go", false,
		"install", fmt.Sprintf("sigs.k8s.io/controller-runtime/tools/setup-envtest@%s", version))
}

// ─────────────────────────────────────────────────────────────────────────────
// controller-gen targets
// ─────────────────────────────────────────────────────────────────────────────

// GenerateManifests runs controller-gen to produce CRD, RBAC, and webhook manifests.
//
//	equivalent to: controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
func (o *OperatorRunner) GenerateManifests() error {
	slog.Info("🔧 Generating CRD/RBAC/webhook manifests...")
	start := time.Now()
	err := o.executor.Run(context.Background(), o.controllerGen(), false,
		`rbac:roleName=manager-role`, `crd`, `webhook`,
		`paths=./...`,
		`output:crd:artifacts:config=config/crd/bases`,
	)
	if err != nil {
		return fmt.Errorf("controller-gen manifests: %w", err)
	}
	slog.Info("✅ Manifests generated", "duration", time.Since(start))
	return nil
}

// GenerateCode runs controller-gen to produce DeepCopy/DeepCopyInto/DeepCopyObject methods.
// When boilerplate is empty, no headerFile argument is passed (bare object generation).
// When boilerplate is non-empty, it is used as the headerFile path.
func (o *OperatorRunner) GenerateCode(boilerplate string) error {
	slog.Info("🔧 Generating DeepCopy code...", "boilerplate", boilerplate)
	start := time.Now()

	args := []string{}
	if boilerplate != "" {
		args = append(args, fmt.Sprintf(`object:headerFile=%s`, boilerplate))
	} else {
		args = append(args, `object`)
	}

	paths := "./..."
	if o.config != nil && o.config.Generate != nil && o.config.Generate.APIPaths != "" {
		paths = o.config.Generate.APIPaths
	}
	args = append(args, fmt.Sprintf(`paths=%s`, paths))

	err := o.executor.Run(context.Background(), o.controllerGen(), false, args...)
	if err != nil {
		return fmt.Errorf("controller-gen generate: %w", err)
	}
	slog.Info("✅ DeepCopy code generated", "duration", time.Since(start))
	return nil
}

// GenerateAll runs DeepCopy generation, manifest generation, and CRD sync in one call.
// It uses config values when available, falling back to supplied parameters.
func (o *OperatorRunner) GenerateAll(boilerplate, helmCRDDir string) error {
	bp := boilerplate
	hcd := helmCRDDir
	if o.config != nil && o.config.Generate != nil {
		if bp == "" {
			bp = o.config.Generate.Boilerplate
		}
		if hcd == "" {
			hcd = o.config.Generate.HelmCRDDir
		}
	}

	if err := o.GenerateCode(bp); err != nil {
		return err
	}
	if hcd != "" {
		return o.GenerateAndSync(hcd)
	}
	return o.GenerateManifests()
}

// GenerateFromConfig runs GenerateAll using loaded config values.
func (o *OperatorRunner) GenerateFromConfig() error {
	if o.config == nil || o.config.Generate == nil {
		return fmt.Errorf("no generate configuration loaded")
	}
	return o.GenerateAll(o.config.Generate.Boilerplate, o.config.Generate.HelmCRDDir)
}

// ─────────────────────────────────────────────────────────────────────────────
// kustomize targets
// ─────────────────────────────────────────────────────────────────────────────

// KustomizeBuild runs `kustomize build <dir>` and returns the rendered YAML as a string.
func (o *OperatorRunner) KustomizeBuild(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("kustomize dir is required")
	}
	slog.Info("📦 Running kustomize build...", "dir", dir)

	tmp, err := os.CreateTemp("", "kustomize-*.yaml")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := o.executor.Run(context.Background(), o.kustomize(), false, "build", dir); err != nil {
		return "", fmt.Errorf("kustomize build %s: %w", dir, err)
	}
	return "", nil
}

// InstallCRDs installs CRDs into the cluster by kustomize-building config/crd and piping to kubectl apply.
func (o *OperatorRunner) InstallCRDs() error {
	slog.Info("📥 Installing CRDs into cluster...")
	start := time.Now()
	if err := o.executor.Run(context.Background(), "sh", false,
		"-c", fmt.Sprintf(`"%s" build config/crd | kubectl apply -f -`, o.kustomize()),
	); err != nil {
		return fmt.Errorf("install CRDs: %w", err)
	}
	slog.Info("✅ CRDs installed", "duration", time.Since(start))
	return nil
}

// UninstallCRDs removes CRDs from the cluster.
func (o *OperatorRunner) UninstallCRDs() error {
	slog.Info("🗑️  Uninstalling CRDs from cluster...")
	start := time.Now()
	if err := o.executor.Run(context.Background(), "sh", false,
		"-c", fmt.Sprintf(`"%s" build config/crd | kubectl delete --ignore-not-found=true -f -`, o.kustomize()),
	); err != nil {
		return fmt.Errorf("uninstall CRDs: %w", err)
	}
	slog.Info("✅ CRDs uninstalled", "duration", time.Since(start))
	return nil
}

// Deploy deploys the operator using kustomize build config/default | kubectl apply.
func (o *OperatorRunner) Deploy(image string) error {
	if image != "" {
		slog.Info("🖼️  Setting manager image", "image", image)
		if err := o.executor.Run(context.Background(), o.kustomize(), false,
			"edit", "set", "image", fmt.Sprintf("controller=%s", image),
		); err != nil {
			return fmt.Errorf("kustomize set image: %w", err)
		}
	}
	slog.Info("🚀 Deploying operator...")
	start := time.Now()
	if err := o.executor.Run(context.Background(), "sh", false,
		"-c", fmt.Sprintf(`"%s" build config/default | kubectl apply -f -`, o.kustomize()),
	); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	slog.Info("✅ Operator deployed", "duration", time.Since(start))
	return nil
}

// Undeploy removes the operator from the cluster.
func (o *OperatorRunner) Undeploy() error {
	slog.Info("🗑️  Undeploying operator...")
	start := time.Now()
	if err := o.executor.Run(context.Background(), "sh", false,
		"-c", fmt.Sprintf(`"%s" build config/default | kubectl delete --ignore-not-found=true -f -`, o.kustomize()),
	); err != nil {
		return fmt.Errorf("undeploy: %w", err)
	}
	slog.Info("✅ Operator undeployed", "duration", time.Since(start))
	return nil
}

// BuildInstaller generates dist/install.yaml — a single YAML with all CRDs + the Deployment.
func (o *OperatorRunner) BuildInstaller(image string) error {
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return fmt.Errorf("mkdir dist: %w", err)
	}
	if image != "" {
		if err := o.executor.Run(context.Background(), o.kustomize(), false,
			"edit", "set", "image", fmt.Sprintf("controller=%s", image),
		); err != nil {
			return fmt.Errorf("kustomize set image: %w", err)
		}
	}
	slog.Info("📦 Building install.yaml...")
	if err := o.executor.Run(context.Background(), "sh", false,
		"-c", fmt.Sprintf(`"%s" build config/default > dist/install.yaml`, o.kustomize()),
	); err != nil {
		return fmt.Errorf("build installer: %w", err)
	}
	slog.Info("✅ dist/install.yaml written")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CRD sync
// ─────────────────────────────────────────────────────────────────────────────

// SyncCRDs copies CRD files from the controller-gen output directory to a Helm chart crds/ directory.
func (o *OperatorRunner) SyncCRDs(srcDir, destDir string) error {
	if srcDir == "" {
		srcDir = "config/crd/bases"
	}
	if destDir == "" {
		return fmt.Errorf("destDir (Helm chart crds/ path) is required")
	}

	slog.Info("🔄 Syncing CRDs to Helm chart...", "src", srcDir, "dst", destDir)
	start := time.Now()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read CRD dir %s: %w", srcDir, err)
	}

	copied := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) < 5 || name[len(name)-5:] != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			return fmt.Errorf("read CRD %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(destDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write CRD %s: %w", name, err)
		}
		copied++
	}

	slog.Info("✅ CRDs synced to Helm chart", "count", copied, "duration", time.Since(start))
	return nil
}

// GenerateAndSync runs GenerateManifests then copies CRDs to the Helm chart.
func (o *OperatorRunner) GenerateAndSync(helmCRDDir string) error {
	if err := o.GenerateManifests(); err != nil {
		return err
	}
	srcDir := "config/crd/bases"
	if o.config != nil && o.config.Generate != nil && o.config.Generate.CRDSrcDir != "" {
		srcDir = o.config.Generate.CRDSrcDir
	}
	return o.SyncCRDs(srcDir, helmCRDDir)
}

// ─────────────────────────────────────────────────────────────────────────────
// Clean
// ─────────────────────────────────────────────────────────────────────────────

// Clean removes build artifacts. When config is loaded, it uses config.Clean.Paths;
// otherwise it removes the supplied paths.
func (o *OperatorRunner) Clean(paths ...string) error {
	targets := paths
	if len(targets) == 0 && o.config != nil && o.config.Clean != nil {
		targets = o.config.Clean.Paths
	}
	if len(targets) == 0 {
		slog.Info("nothing to clean")
		return nil
	}

	slog.Info("🧹 Cleaning build artifacts...", "paths", targets)
	for _, p := range targets {
		if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clean %s: %w", p, err)
		}
	}
	slog.Info("✅ Clean complete")
	return nil
}

// CleanFromConfig removes build artifacts using loaded config.
func (o *OperatorRunner) CleanFromConfig() error {
	if o.config == nil || o.config.Clean == nil {
		return fmt.Errorf("no clean configuration loaded")
	}
	return o.Clean()
}

// ─────────────────────────────────────────────────────────────────────────────
// setup-envtest
// ─────────────────────────────────────────────────────────────────────────────

// EnvtestAssets downloads the envtest binaries for the given Kubernetes version
// (e.g. "1.36") and returns the path that should be set as KUBEBUILDER_ASSETS.
func (o *OperatorRunner) EnvtestAssets(k8sVersion string) (string, error) {
	if k8sVersion == "" {
		k8sVersion = "1.36"
	}
	slog.Info("⬇️  Setting up envtest assets", "k8sVersion", k8sVersion)

	tmp, err := os.CreateTemp("", "envtest-path-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := o.executor.Run(context.Background(), "sh", false,
		"-c",
		fmt.Sprintf(`"%s" use %s --bin-dir "%s" -p path > "%s"`,
			o.envtest(), k8sVersion, o.LocalBin, tmp.Name()),
	); err != nil {
		return "", fmt.Errorf("setup-envtest: %w", err)
	}
	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", err
	}
	path := string(data)
	slog.Info("✅ Envtest assets ready", "KUBEBUILDER_ASSETS", path)
	return path, nil
}
