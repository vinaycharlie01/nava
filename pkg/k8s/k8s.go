package k8sx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nirantaraai/nava/pkg/exec"
)

// K8sRunner handles kubectl command execution with dependency injection
type K8sRunner struct {
	executor execx.Executor
}

// NewK8sRunner creates a new K8sRunner with the default executor
func NewK8sRunner() *K8sRunner {
	return &K8sRunner{
		executor: execx.NewExec(),
	}
}

// NewK8sRunnerWithExecutor creates a new K8sRunner with a custom executor
func NewK8sRunnerWithExecutor(executor execx.Executor) *K8sRunner {
	return &K8sRunner{
		executor: executor,
	}
}

// ApplyOptions contains options for kubectl apply
type ApplyOptions struct {
	Filenames  []string
	Namespace  string
	Kustomize  string
	ServerSide bool
	Force      bool
	Stdin      string // If provided, used as -f -
}

// Apply applies a configuration to a resource by filename or stdin
func (k *K8sRunner) Apply(opts ApplyOptions) error {
	if len(opts.Filenames) == 0 && opts.Kustomize == "" && opts.Stdin == "" {
		return fmt.Errorf("either filenames, kustomize path, or stdin is required")
	}

	slog.Info("🚀 Applying Kubernetes manifests...",
		"filenames", opts.Filenames,
		"kustomize", opts.Kustomize,
		"namespace", opts.Namespace,
		"stdin_len", len(opts.Stdin),
	)

	start := time.Now()

	args := []string{"apply"}

	if opts.Kustomize != "" {
		args = append(args, "-k", opts.Kustomize)
	} else if opts.Stdin != "" {
		// Use temp file for stdin content
		f, err := os.CreateTemp("", "k8sx-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(f.Name())

		if _, err := f.WriteString(opts.Stdin); err != nil {
			f.Close()
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		f.Close()
		args = append(args, "-f", f.Name())
	} else {
		for _, f := range opts.Filenames {
			args = append(args, "-f", f)
		}
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.ServerSide {
		args = append(args, "--server-side")
	}

	if opts.Force {
		args = append(args, "--force")
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes manifests applied", "duration", time.Since(start))
	return nil
}

// DeleteOptions contains options for kubectl delete
type DeleteOptions struct {
	Filenames []string
	Namespace string
	Kustomize string
	Wait      bool
	Force     bool
}

// Delete deletes resources by filenames, stdin, resources and names, or by resources and label selector
func (k *K8sRunner) Delete(opts DeleteOptions, resourceNames ...string) error {
	slog.Info("🗑️  Deleting Kubernetes resources...",
		"filenames", opts.Filenames,
		"kustomize", opts.Kustomize,
		"namespace", opts.Namespace,
		"names", resourceNames,
	)

	start := time.Now()

	args := []string{"delete"}

	if opts.Kustomize != "" {
		args = append(args, "-k", opts.Kustomize)
	} else if len(opts.Filenames) > 0 {
		for _, f := range opts.Filenames {
			args = append(args, "-f", f)
		}
	} else if len(resourceNames) > 0 {
		// If no filenames/kustomize, we assume resourceType and names are provided elsewhere or in resourceNames
		// Actually, let's keep it simple: if resourceNames is provided, we use it.
		// We might need a resourceType though. Let's assume the first element of resourceNames is the type if no filenames.
		args = append(args, resourceNames...)
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if !opts.Wait {
		args = append(args, "--wait=false")
	}

	if opts.Force {
		args = append(args, "--force", "--grace-period=0")
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes resources deleted", "duration", time.Since(start))
	return nil
}

// CreateNamespace creates a namespace
func (k *K8sRunner) CreateNamespace(name string) error {
	if name == "" {
		return fmt.Errorf("namespace name is required")
	}

	slog.Info("📁 Creating Kubernetes namespace...", "name", name)

	start := time.Now()

	args := []string{"create", "namespace", name}

	err := k.executor.Run(context.Background(), "kubectl", false, args...)
	if err != nil {
		// If it already exists, we don't want to return an error
		slog.Warn("⚠️  Namespace might already exist", "name", name, "error", err)
		return nil
	}

	slog.Info("✅ Kubernetes namespace created", "name", name, "duration", time.Since(start))
	return nil
}

// Get displays one or many resources
func (k *K8sRunner) Get(resourceType, name, namespace string, args ...string) error {
	slog.Info("🔍 Getting Kubernetes resources...",
		"type", resourceType,
		"name", name,
		"namespace", namespace,
	)

	start := time.Now()

	cmdArgs := []string{"get", resourceType}
	if name != "" {
		cmdArgs = append(cmdArgs, name)
	}

	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	} else {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	}

	cmdArgs = append(cmdArgs, args...)

	if err := k.executor.Run(context.Background(), "kubectl", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes resources retrieved", "duration", time.Since(start))
	return nil
}

// WaitOptions contains options for kubectl wait
type WaitOptions struct {
	Resource  string
	For       string
	Namespace string
	Timeout   string
	Selector  string
}

// Wait waits for a specific condition on one or many resources
func (k *K8sRunner) Wait(opts WaitOptions) error {
	if opts.Resource == "" && opts.Selector == "" {
		return fmt.Errorf("resource or selector is required")
	}
	if opts.For == "" {
		return fmt.Errorf("condition (for) is required")
	}

	slog.Info("⏳ Waiting for Kubernetes resource condition...",
		"resource", opts.Resource,
		"selector", opts.Selector,
		"condition", opts.For,
		"namespace", opts.Namespace,
	)

	start := time.Now()

	args := []string{"wait"}

	if opts.Resource != "" {
		args = append(args, opts.Resource)
	}

	if opts.Selector != "" {
		args = append(args, "--selector", opts.Selector)
	}

	args = append(args, "--for", opts.For)

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout", opts.Timeout)
	} else {
		args = append(args, "--timeout", "300s")
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes resource condition met", "duration", time.Since(start))
	return nil
}

// RolloutStatus displays the status of a rollout
func (k *K8sRunner) RolloutStatus(resource, namespace string, timeout string) error {
	if resource == "" {
		return fmt.Errorf("resource is required")
	}

	slog.Info("📊 Checking rollout status...", "resource", resource, "namespace", namespace)

	start := time.Now()

	args := []string{"rollout", "status", resource}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if timeout != "" {
		args = append(args, "--timeout", timeout)
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Rollout complete", "duration", time.Since(start))
	return nil
}

// CreateSecretOptions contains options for kubectl create secret
type CreateSecretOptions struct {
	Name        string
	Namespace   string
	Type        string // generic, docker-registry, tls
	FromLiteral []string
	FromFile    []string
}

// CreateSecret creates a secret
func (k *K8sRunner) CreateSecret(opts CreateSecretOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("secret name is required")
	}

	slog.Info("🔑 Creating Kubernetes secret...", "name", opts.Name, "namespace", opts.Namespace)

	start := time.Now()

	args := []string{"create", "secret", opts.Type, opts.Name}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	for _, l := range opts.FromLiteral {
		args = append(args, "--from-literal="+l)
	}

	for _, f := range opts.FromFile {
		args = append(args, "--from-file="+f)
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes secret created", "duration", time.Since(start))
	return nil
}

// Label labels a resource
func (k *K8sRunner) Label(resourceType, name, namespace string, labels map[string]string, overwrite bool) error {
	if resourceType == "" || name == "" {
		return fmt.Errorf("resource type and name are required")
	}

	slog.Info("🏷️  Labeling Kubernetes resource...",
		"type", resourceType,
		"name", name,
		"namespace", namespace,
		"labels", labels,
	)

	start := time.Now()

	args := []string{"label", resourceType, name}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	for k, v := range labels {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}

	if overwrite {
		args = append(args, "--overwrite")
	}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes resource labeled", "duration", time.Since(start))
	return nil
}

// Patch patches a resource
func (k *K8sRunner) Patch(resourceType, name, namespace, patchType, patch string) error {
	if resourceType == "" || name == "" || patch == "" {
		return fmt.Errorf("resource type, name, and patch are required")
	}

	slog.Info("🩹 Patching Kubernetes resource...",
		"type", resourceType,
		"name", name,
		"namespace", namespace,
		"patchType", patchType,
	)

	start := time.Now()

	args := []string{"patch", resourceType, name}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if patchType != "" {
		args = append(args, "--type", patchType)
	}

	args = append(args, "-p", patch)

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes resource patched", "duration", time.Since(start))
	return nil
}

// Run executes a raw kubectl command
func (k *K8sRunner) Run(args ...string) error {
	slog.Info("☸️  Running kubectl command", "args", args)
	return k.executor.Run(context.Background(), "kubectl", false, args...)
}

// ConfigUseContext sets the current context in the kubeconfig file
func (k *K8sRunner) ConfigUseContext(contextName string) error {
	if contextName == "" {
		return fmt.Errorf("context name is required")
	}

	slog.Info("☸️  Switching Kubernetes context...", "context", contextName)

	start := time.Now()

	args := []string{"config", "use-context", contextName}

	if err := k.executor.Run(context.Background(), "kubectl", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Kubernetes context switched", "duration", time.Since(start))
	return nil
}
