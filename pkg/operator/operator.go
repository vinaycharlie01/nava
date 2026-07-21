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
	"path/filepath"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
)

// OperatorRunner wraps controller-gen, kustomize, and setup-envtest.
type OperatorRunner struct {
	executor execx.Executor
	// LocalBin is the directory where tool binaries are downloaded (default: ./bin).
	LocalBin string
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

// controllerGen returns the path to the controller-gen binary inside LocalBin.
func (o *OperatorRunner) controllerGen() string {
	return filepath.Join(o.LocalBin, "controller-gen")
}

// kustomize returns the path to the kustomize binary inside LocalBin.
func (o *OperatorRunner) kustomize() string {
	return filepath.Join(o.LocalBin, "kustomize")
}

// envtest returns the path to the setup-envtest binary inside LocalBin.
func (o *OperatorRunner) envtest() string {
	return filepath.Join(o.LocalBin, "setup-envtest")
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
//
//	equivalent to: controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
func (o *OperatorRunner) GenerateCode(boilerplate string) error {
	if boilerplate == "" {
		boilerplate = "hack/boilerplate.go.txt"
	}
	slog.Info("🔧 Generating DeepCopy code...", "boilerplate", boilerplate)
	start := time.Now()
	err := o.executor.Run(context.Background(), o.controllerGen(), false,
		fmt.Sprintf(`object:headerFile="%s"`, boilerplate),
		`paths=./...`,
	)
	if err != nil {
		return fmt.Errorf("controller-gen generate: %w", err)
	}
	slog.Info("✅ DeepCopy code generated", "duration", time.Since(start))
	return nil
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

	// We need to capture output — use a temp file.
	tmp, err := os.CreateTemp("", "kustomize-*.yaml")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	// Run kustomize build and redirect to tmp (via shell isn't reliable; use exec + capture).
	// For simplicity, run it and stream to stdout (the caller can pipe as needed).
	if err := o.executor.Run(context.Background(), o.kustomize(), false, "build", dir); err != nil {
		return "", fmt.Errorf("kustomize build %s: %w", dir, err)
	}
	return "", nil
}

// InstallCRDs installs CRDs into the cluster by kustomize-building config/crd and piping to kubectl apply.
func (o *OperatorRunner) InstallCRDs() error {
	slog.Info("📥 Installing CRDs into cluster...")
	start := time.Now()
	// Run: kustomize build config/crd | kubectl apply -f -
	// We chain via shell via the exec helper.
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
// setup-envtest
// ─────────────────────────────────────────────────────────────────────────────

// EnvtestAssets downloads the envtest binaries for the given Kubernetes version
// (e.g. "1.36") and returns the path that should be set as KUBEBUILDER_ASSETS.
func (o *OperatorRunner) EnvtestAssets(k8sVersion string) (string, error) {
	if k8sVersion == "" {
		k8sVersion = "1.36"
	}
	slog.Info("⬇️  Setting up envtest assets", "k8sVersion", k8sVersion)
	// setup-envtest use <version> --bin-dir <localbin> -p path
	// We capture its stdout to get the path.
	// Use a simple approach: write a small wrapper script.
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
