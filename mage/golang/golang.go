package golangmagex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// GoRunner handles Go command execution with dependency injection
type GoRunner struct {
	executor execx.Executor
	config   *GoConfig
}

// NewGoRunner creates a new GoRunner with the default executor
func NewGoRunner() *GoRunner {
	return &GoRunner{
		executor: execx.NewExec(),
	}
}

// NewGoRunnerWithExecutor creates a new GoRunner with a custom executor
func NewGoRunnerWithExecutor(executor execx.Executor) *GoRunner {
	return &GoRunner{
		executor: executor,
	}
}

// NewGoRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewGoRunnerFromYAML(filepath string) (*GoRunner, error) {
	runner := NewGoRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads Go configuration from a YAML file
func (g *GoRunner) LoadConfig(filepath string) error {
	config, err := LoadGoConfig(filepath)
	if err != nil {
		return err
	}
	g.config = config
	return nil
}

// GoConfig contains all Go operation configurations
type GoConfig struct {
	Directory string         `yaml:"directory,omitempty"`
	Setup     *SetupOptions  `yaml:"setup,omitempty"`
	Build     *BuildConfig   `yaml:"build,omitempty"`
	Run       *RunConfig     `yaml:"run,omitempty"`
	Test      *TestConfig    `yaml:"test,omitempty"`
	Lint      *LintConfig    `yaml:"lint,omitempty"`
	Vet       *VetConfig     `yaml:"vet,omitempty"`
	Format    *FormatConfig  `yaml:"format,omitempty"`
	Install   *InstallConfig `yaml:"install,omitempty"`
}

// SetupOptions contains options for setting up Go environment
type SetupOptions struct {
	ModDownload bool `yaml:"modDownload,omitempty"`
	ModTidy     bool `yaml:"modTidy,omitempty"`
}

// BuildConfig contains options for building Go binaries
type BuildConfig struct {
	Output string   `yaml:"output"`
	Main   string   `yaml:"main"`
	Args   []string `yaml:"args,omitempty"`
}

// RunConfig contains options for running Go programs
type RunConfig struct {
	Main string   `yaml:"main"`
	Args []string `yaml:"args,omitempty"`
}

// TestConfig contains options for running Go tests
type TestConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Args     []string `yaml:"args,omitempty"`
}

// LintConfig contains options for running golangci-lint
type LintConfig struct {
	Args []string `yaml:"args,omitempty"`
}

// VetConfig contains options for running go vet
type VetConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Args     []string `yaml:"args,omitempty"`
}

// FormatConfig contains options for formatting Go code
type FormatConfig struct {
	Args []string `yaml:"args,omitempty"`
}

// InstallConfig contains options for installing Go packages
type InstallConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Args     []string `yaml:"args,omitempty"`
}

// LoadGoConfig loads Go configuration from a YAML file
func LoadGoConfig(filepath string) (*GoConfig, error) {
	var config GoConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// RunInDir runs a Go command in a specific directory
func (g *GoRunner) RunInDir(dir, command string, args ...string) error {
	slog.Info("🔧 Running Go command in directory...", "dir", dir, "command", command)
	start := time.Now()

	cmdArgs := append([]string{command}, args...)
	if err := g.executor.RunInDir(context.Background(), dir, "go", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("✅ Command completed", "duration", time.Since(start))
	return nil
}

// SetupFromConfig sets up Go environment using loaded config
func (g *GoRunner) SetupFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Setup == nil {
		return fmt.Errorf("no setup configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🎯 Setting up Go environment from config...", "directory", dir)

	if g.config.Setup.ModDownload {
		if err := g.RunInDir(dir, "mod", "download"); err != nil {
			return err
		}
	}

	if g.config.Setup.ModTidy {
		if err := g.RunInDir(dir, "mod", "tidy"); err != nil {
			return err
		}
	}

	slog.Info("✅ Go environment setup complete")
	return nil
}

// BuildFromConfig builds Go binary using loaded config
func (g *GoRunner) BuildFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Build == nil {
		return fmt.Errorf("no build configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🔨 Building Go binary from config...", "directory", dir)

	buildArgs := append([]string{"build", "-o", g.config.Build.Output, g.config.Build.Main}, g.config.Build.Args...)
	if err := g.RunInDir(dir, "build", buildArgs[1:]...); err != nil {
		return err
	}

	slog.Info("✅ Build complete")
	return nil
}

// RunFromConfig runs Go program using loaded config with graceful shutdown
func (g *GoRunner) RunFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Run == nil {
		return fmt.Errorf("no run configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals in goroutine
	go func() {
		sig := <-sigChan
		slog.Info("🛑 Received shutdown signal", "signal", sig)
		slog.Info("⏳ Initiating graceful shutdown...")
		cancel()
	}()

	slog.Info("🚀 Running Go program from config...", "directory", dir)

	runArgs := append([]string{"run", g.config.Run.Main}, g.config.Run.Args...)
	cmdArgs := append([]string{"run"}, runArgs[1:]...)

	if err := g.executor.RunInDir(ctx, dir, "go", false, cmdArgs...); err != nil {
		// Check if error is due to context cancellation (graceful shutdown)
		if errors.Is(err, context.Canceled) {
			slog.Info("✅ Program stopped gracefully")
			return nil
		}
		return err
	}

	return nil
}

// TestFromConfig runs Go tests using loaded config
func (g *GoRunner) TestFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Test == nil {
		return fmt.Errorf("no test configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := g.config.Test.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	slog.Info("🧪 Running Go tests from config...", "directory", dir)
	start := time.Now()

	testArgs := append([]string{"test"}, packages...)
	testArgs = append(testArgs, g.config.Test.Args...)

	if err := g.RunInDir(dir, "test", testArgs[1:]...); err != nil {
		return err
	}

	slog.Info("✅ Tests passed", "duration", time.Since(start))
	return nil
}

// LintFromConfig runs golangci-lint using loaded config
func (g *GoRunner) LintFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Lint == nil {
		return fmt.Errorf("no lint configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🔍 Running golangci-lint from config...", "directory", dir)
	start := time.Now()

	defaultArgs := []string{"run", "--timeout=5m"}
	lintArgs := append(defaultArgs, g.config.Lint.Args...)

	if err := g.executor.RunInDir(context.Background(), dir, "golangci-lint", false, lintArgs...); err != nil {
		return err
	}

	slog.Info("✅ Linting passed", "duration", time.Since(start))
	return nil
}

// VetFromConfig runs go vet using loaded config
func (g *GoRunner) VetFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Vet == nil {
		return fmt.Errorf("no vet configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := g.config.Vet.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	slog.Info("🔍 Running go vet from config...", "directory", dir)
	start := time.Now()

	vetArgs := append([]string{"vet"}, packages...)
	vetArgs = append(vetArgs, g.config.Vet.Args...)

	if err := g.RunInDir(dir, "vet", vetArgs[1:]...); err != nil {
		return err
	}

	slog.Info("✅ Go vet passed", "duration", time.Since(start))
	return nil
}

// FormatFromConfig formats Go code using loaded config
func (g *GoRunner) FormatFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Format == nil {
		return fmt.Errorf("no format configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("✨ Formatting Go code from config...", "directory", dir)
	start := time.Now()

	defaultArgs := []string{"-w", "."}
	formatArgs := append(defaultArgs, g.config.Format.Args...)

	if err := g.executor.RunInDir(context.Background(), dir, "gofmt", false, formatArgs...); err != nil {
		return err
	}

	slog.Info("✅ Formatting complete", "duration", time.Since(start))
	return nil
}

// InstallFromConfig installs Go packages using loaded config
func (g *GoRunner) InstallFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Install == nil {
		return fmt.Errorf("no install configuration found")
	}

	if len(g.config.Install.Packages) == 0 {
		slog.Info("ℹ️  No packages to install")
		return nil
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("📦 Installing Go packages from config...", "directory", dir, "packages", g.config.Install.Packages)
	start := time.Now()

	for _, pkg := range g.config.Install.Packages {
		installArgs := append([]string{"install", pkg}, g.config.Install.Args...)
		if err := g.RunInDir(dir, "install", installArgs[1:]...); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}

	slog.Info("✅ Installation complete", "duration", time.Since(start))
	return nil
}

// Package-level convenience functions for mage targets
var defaultRunner = NewGoRunner()

// LoadConfig loads Go configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*GoRunner, error) {
	return NewGoRunnerFromYAML(filepath)
}

// Setup sets up Go environment (requires loaded config)
func Setup() error {
	return defaultRunner.SetupFromConfig()
}

// Build builds Go binary (requires loaded config)
func Build() error {
	return defaultRunner.BuildFromConfig()
}

// Run runs Go program (requires loaded config)
func Run() error {
	return defaultRunner.RunFromConfig()
}

// Test runs Go tests (requires loaded config)
func Test() error {
	return defaultRunner.TestFromConfig()
}

// Lint runs golangci-lint (requires loaded config)
func Lint() error {
	return defaultRunner.LintFromConfig()
}

// Vet runs go vet (requires loaded config)
func Vet() error {
	return defaultRunner.VetFromConfig()
}

// Format formats Go code (requires loaded config)
func Format() error {
	return defaultRunner.FormatFromConfig()
}

// Install installs Go packages (requires loaded config)
func Install() error {
	return defaultRunner.InstallFromConfig()
}
