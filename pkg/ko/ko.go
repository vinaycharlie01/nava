package kox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// KoRunner handles ko command execution with dependency injection
type KoRunner struct {
	executor execx.Executor
	config   *KoConfig
}

// NewKoRunner creates a new KoRunner with the default executor
func NewKoRunner() *KoRunner {
	return &KoRunner{
		executor: execx.NewExec(),
	}
}

// NewKoRunnerWithExecutor creates a new KoRunner with a custom executor
func NewKoRunnerWithExecutor(executor execx.Executor) *KoRunner {
	return &KoRunner{
		executor: executor,
	}
}

// NewKoRunnerFromYAML creates a new KoRunner with configuration loaded from YAML
func NewKoRunnerFromYAML(filepath string) (*KoRunner, error) {
	config, err := LoadKoConfig(filepath)
	if err != nil {
		return nil, err
	}
	return &KoRunner{
		executor: execx.NewExec(),
		config:   config,
	}, nil
}

// LoadConfig loads ko configuration from a YAML file
func (k *KoRunner) LoadConfig(filepath string) error {
	config, err := LoadKoConfig(filepath)
	if err != nil {
		return err
	}
	k.config = config
	return nil
}

// KoConfig contains all ko operation configurations
type KoConfig struct {
	Build   *BuildOptions   `yaml:"build,omitempty"`
	Apply   *ApplyOptions   `yaml:"apply,omitempty"`
	Delete  *DeleteOptions  `yaml:"delete,omitempty"`
	Resolve *ResolveOptions `yaml:"resolve,omitempty"`
	Publish *PublishOptions `yaml:"publish,omitempty"`
}

// LoadKoConfig loads ko configuration from a YAML file
func LoadKoConfig(filepath string) (*KoConfig, error) {
	var config KoConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// BuildOptions contains options for ko build
type BuildOptions struct {
	ImportPath          string   `yaml:"importPath"`                    // Go import path to build
	Tags                []string `yaml:"tags,omitempty"`                // Image tags
	Platform            []string `yaml:"platform,omitempty"`            // Target platforms (e.g., linux/amd64,linux/arm64)
	BaseImage           string   `yaml:"baseImage,omitempty"`           // Base image to use
	Bare                bool     `yaml:"bare,omitempty"`                // Whether to use a bare image
	Local               bool     `yaml:"local,omitempty"`               // Build locally without pushing
	Push                bool     `yaml:"push,omitempty"`                // Push to registry
	PreserveImportPaths bool     `yaml:"preserveImportPaths,omitempty"` // Preserve import paths in image names
}

// Build builds a container image using ko with loaded configuration
func (k *KoRunner) Build() error {
	if k.config == nil || k.config.Build == nil {
		return fmt.Errorf("build configuration not loaded")
	}

	opts := k.config.Build
	if opts.ImportPath == "" {
		return fmt.Errorf("import path is required")
	}

	slog.Info("🐳 Building container image with ko...",
		"importPath", opts.ImportPath,
		"local", opts.Local,
		"push", opts.Push,
	)

	start := time.Now()

	args := []string{"build", opts.ImportPath}

	for _, tag := range opts.Tags {
		args = append(args, "--tags", tag)
	}

	for _, platform := range opts.Platform {
		args = append(args, "--platform", platform)
	}

	if opts.BaseImage != "" {
		args = append(args, "--base-import-paths", opts.BaseImage)
	}

	if opts.Bare {
		args = append(args, "--bare")
	}

	if opts.Local {
		args = append(args, "--local")
	}

	if opts.Push {
		args = append(args, "--push")
	}

	if opts.PreserveImportPaths {
		args = append(args, "--preserve-import-paths")
	}

	if err := k.executor.Run(context.Background(), "ko", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Container image built", "duration", time.Since(start))
	return nil
}

// ApplyOptions contains options for ko apply
type ApplyOptions struct {
	Filenames           []string `yaml:"filenames"`                     // Kubernetes manifest files
	Recursive           bool     `yaml:"recursive,omitempty"`           // Process directories recursively
	Selector            string   `yaml:"selector,omitempty"`            // Label selector
	BaseImage           string   `yaml:"baseImage,omitempty"`           // Base image to use
	Platform            []string `yaml:"platform,omitempty"`            // Target platforms
	Local               bool     `yaml:"local,omitempty"`               // Build locally without pushing
	Bare                bool     `yaml:"bare,omitempty"`                // Use bare image
	PreserveImportPaths bool     `yaml:"preserveImportPaths,omitempty"` // Preserve import paths
}

// Apply builds images and applies Kubernetes manifests using loaded configuration
func (k *KoRunner) Apply() error {
	if k.config == nil || k.config.Apply == nil {
		return fmt.Errorf("apply configuration not loaded")
	}

	opts := k.config.Apply
	if len(opts.Filenames) == 0 {
		return fmt.Errorf("at least one filename is required")
	}

	slog.Info("🚀 Building and applying with ko...",
		"files", opts.Filenames,
		"local", opts.Local,
	)

	start := time.Now()

	args := []string{"apply"}

	for _, filename := range opts.Filenames {
		args = append(args, "-f", filename)
	}

	if opts.Recursive {
		args = append(args, "--recursive")
	}

	if opts.Selector != "" {
		args = append(args, "--selector", opts.Selector)
	}

	if opts.BaseImage != "" {
		args = append(args, "--base-import-paths", opts.BaseImage)
	}

	for _, platform := range opts.Platform {
		args = append(args, "--platform", platform)
	}

	if opts.Local {
		args = append(args, "--local")
	}

	if opts.Bare {
		args = append(args, "--bare")
	}

	if opts.PreserveImportPaths {
		args = append(args, "--preserve-import-paths")
	}

	if err := k.executor.Run(context.Background(), "ko", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Images built and manifests applied", "duration", time.Since(start))
	return nil
}

// DeleteOptions contains options for ko delete
type DeleteOptions struct {
	Filenames []string `yaml:"filenames"`           // Kubernetes manifest files
	Recursive bool     `yaml:"recursive,omitempty"` // Process directories recursively
	Selector  string   `yaml:"selector,omitempty"`  // Label selector
}

// Delete deletes Kubernetes resources using loaded configuration
func (k *KoRunner) Delete() error {
	if k.config == nil || k.config.Delete == nil {
		return fmt.Errorf("delete configuration not loaded")
	}

	opts := k.config.Delete
	if len(opts.Filenames) == 0 {
		return fmt.Errorf("at least one filename is required")
	}

	slog.Info("🗑️  Deleting resources with ko...", "files", opts.Filenames)

	start := time.Now()

	args := []string{"delete"}

	for _, filename := range opts.Filenames {
		args = append(args, "-f", filename)
	}

	if opts.Recursive {
		args = append(args, "--recursive")
	}

	if opts.Selector != "" {
		args = append(args, "--selector", opts.Selector)
	}

	if err := k.executor.Run(context.Background(), "ko", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Resources deleted", "duration", time.Since(start))
	return nil
}

// ResolveOptions contains options for ko resolve
type ResolveOptions struct {
	ImportPaths []string `yaml:"importPaths"`    // Import paths to resolve
	Args        []string `yaml:"args,omitempty"` // Additional arguments
}

// Resolve resolves import paths to image references using loaded configuration
func (k *KoRunner) Resolve() error {
	if k.config == nil || k.config.Resolve == nil {
		return fmt.Errorf("resolve configuration not loaded")
	}

	opts := k.config.Resolve
	if len(opts.ImportPaths) == 0 {
		return fmt.Errorf("at least one import path is required")
	}

	slog.Info("🔍 Resolving import paths...", "paths", opts.ImportPaths)

	start := time.Now()

	cmdArgs := []string{"resolve"}
	cmdArgs = append(cmdArgs, opts.Args...)
	cmdArgs = append(cmdArgs, opts.ImportPaths...)

	if err := k.executor.Run(context.Background(), "ko", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Import paths resolved", "duration", time.Since(start))
	return nil
}

// PublishOptions contains options for ko publish
type PublishOptions struct {
	ImportPath string   `yaml:"importPath"`     // Import path to publish
	Args       []string `yaml:"args,omitempty"` // Additional arguments
}

// Publish publishes images for import paths using loaded configuration
func (k *KoRunner) Publish() error {
	if k.config == nil || k.config.Publish == nil {
		return fmt.Errorf("publish configuration not loaded")
	}

	opts := k.config.Publish
	if opts.ImportPath == "" {
		return fmt.Errorf("import path is required")
	}

	slog.Info("📤 Publishing image...", "importPath", opts.ImportPath)

	start := time.Now()

	cmdArgs := []string{"publish", opts.ImportPath}
	cmdArgs = append(cmdArgs, opts.Args...)

	if err := k.executor.Run(context.Background(), "ko", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Image published", "duration", time.Since(start))
	return nil
}

// Made with Bob
