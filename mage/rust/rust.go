package rustmagex

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// RustRunner handles Rust command execution with dependency injection
type RustRunner struct {
	executor execx.Executor
	config   *RustConfig
}

// NewRustRunner creates a new RustRunner with the default executor
func NewRustRunner() *RustRunner {
	return &RustRunner{
		executor: execx.NewExec(),
	}
}

// NewRustRunnerWithExecutor creates a new RustRunner with a custom executor
func NewRustRunnerWithExecutor(executor execx.Executor) *RustRunner {
	return &RustRunner{
		executor: executor,
	}
}

// NewRustRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRustRunnerFromYAML(filepath string) (*RustRunner, error) {
	runner := NewRustRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads Rust configuration from a YAML file
func (r *RustRunner) LoadConfig(filepath string) error {
	config, err := LoadRustConfig(filepath)
	if err != nil {
		return err
	}
	r.config = config
	return nil
}

// RustConfig contains all Rust operation configurations
type RustConfig struct {
	Directory   string               `yaml:"directory,omitempty"`
	GenCommands []*GenCommandConfig  `yaml:"genCommands,omitempty"`
	Setup       *SetupOptions        `yaml:"setup,omitempty"`
	Build       *CargoCommandOptions `yaml:"build,omitempty"`
	Run         *CargoCommandOptions `yaml:"run,omitempty"`
	Test        *CargoCommandOptions `yaml:"test,omitempty"`
	Lint        *LintOptions         `yaml:"lint,omitempty"`
	Format      *CargoCommandOptions `yaml:"format,omitempty"`
	Clean       *CleanConfig         `yaml:"clean,omitempty"`
}

// GenCommandConfig contains configuration for a named cargo command
type GenCommandConfig struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,omitempty"`
}

// SetupOptions contains options for setting up Rust environment
type SetupOptions struct {
	CheckToolchain bool `yaml:"checkToolchain,omitempty"`
	Fetch          bool `yaml:"fetch,omitempty"`
}

// CargoCommandOptions contains command args for cargo subcommands
type CargoCommandOptions struct {
	Args []string `yaml:"args,omitempty"`
}

// LintOptions contains options for running Rust linting
type LintOptions struct {
	Tool string   `yaml:"tool,omitempty"`
	Args []string `yaml:"args,omitempty"`
}

// CleanConfig contains options for cleaning Rust project
type CleanConfig struct {
	UseCargoClean bool     `yaml:"useCargoClean,omitempty"`
	Paths         []string `yaml:"paths,omitempty"`
}

// LoadRustConfig loads Rust configuration from a YAML file
func LoadRustConfig(filepath string) (*RustConfig, error) {
	var config RustConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

func resolveDir(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

// EnsureCargoProject validates that Cargo.toml exists in the target directory.
func (r *RustRunner) EnsureCargoProject(dir string) error {
	cargoToml := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(cargoToml); err != nil {
		return fmt.Errorf("Cargo.toml not found in %s", dir)
	}
	return nil
}

// RunCargoCommand runs a cargo command in a specific directory.
func (r *RustRunner) RunCargoCommand(dir, command string, args ...string) error {
	dir = resolveDir(dir)

	slog.Info("🦀 Running cargo command...", "directory", dir, "command", command)
	start := time.Now()

	cmdArgs := append([]string{command}, args...)
	if err := r.executor.RunInDir(context.Background(), dir, "cargo", false, cmdArgs...); err != nil {
		return fmt.Errorf("cargo %s failed: %w", command, err)
	}

	slog.Info("✅ Cargo command completed", "duration", time.Since(start))
	return nil
}

// CheckToolchain verifies that rustc and cargo are installed.
func (r *RustRunner) CheckToolchain() error {
	slog.Info("🔎 Verifying Rust toolchain...")

	if err := r.executor.Run(context.Background(), "rustc", false, "--version"); err != nil {
		return fmt.Errorf("failed to verify rustc installation: %w", err)
	}

	if err := r.executor.Run(context.Background(), "cargo", false, "--version"); err != nil {
		return fmt.Errorf("failed to verify cargo installation: %w", err)
	}

	slog.Info("✅ Rust toolchain verified")
	return nil
}

// FetchDependencies fetches Rust dependencies for the project.
func (r *RustRunner) FetchDependencies(dir string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	return r.RunCargoCommand(dir, "fetch", args...)
}

// BuildProject builds a Rust project.
func (r *RustRunner) BuildProject(dir string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	return r.RunCargoCommand(dir, "build", args...)
}

// RunProject runs a Rust project.
func (r *RustRunner) RunProject(dir string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	return r.RunCargoCommand(dir, "run", args...)
}

// TestProject runs Rust tests.
func (r *RustRunner) TestProject(dir string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	return r.RunCargoCommand(dir, "test", args...)
}

// LintProject runs Rust linting using clippy by default.
func (r *RustRunner) LintProject(dir, tool string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	if tool == "" {
		tool = "clippy"
	}

	return r.RunCargoCommand(dir, tool, args...)
}

// FormatProject formats Rust code using cargo fmt.
func (r *RustRunner) FormatProject(dir string, args ...string) error {
	dir = resolveDir(dir)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	return r.RunCargoCommand(dir, "fmt", args...)
}

// Setup sets up a Rust environment (toolchain check + dependency fetch).
func (r *RustRunner) Setup(dir string, checkToolchain bool, fetch bool) error {
	dir = resolveDir(dir)

	slog.Info("🎯 Setting up Rust environment...", "directory", dir)

	if checkToolchain {
		if err := r.CheckToolchain(); err != nil {
			return err
		}
	}

	if fetch {
		if err := r.FetchDependencies(dir); err != nil {
			return err
		}
	}

	slog.Info("✅ Rust environment setup complete")
	return nil
}

// CleanProject cleans Rust project artifacts.
func (r *RustRunner) CleanProject(dir string, useCargoClean bool, paths ...string) error {
	dir = resolveDir(dir)

	slog.Info("🧹 Cleaning Rust project...", "directory", dir)

	if useCargoClean {
		if err := r.RunCargoCommand(dir, "clean"); err != nil {
			return err
		}
	}

	if len(paths) == 0 {
		paths = []string{"target"}
	}

	for _, path := range paths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			slog.Info("Removing directory/file", "path", fullPath)
			if err := os.RemoveAll(fullPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", fullPath, err)
			}
		}
	}

	slog.Info("✅ Clean complete")
	return nil
}

// SetupFromConfig sets up Rust environment using loaded config.
func (r *RustRunner) SetupFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Setup == nil {
		return fmt.Errorf("no setup configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.Setup(dir, r.config.Setup.CheckToolchain, r.config.Setup.Fetch)
}

// BuildFromConfig builds Rust project using loaded config.
func (r *RustRunner) BuildFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Build == nil {
		return fmt.Errorf("no build configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.BuildProject(dir, r.config.Build.Args...)
}

// RunFromConfig runs Rust application using loaded config.
func (r *RustRunner) RunFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Run == nil {
		return fmt.Errorf("no run configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.RunProject(dir, r.config.Run.Args...)
}

// TestFromConfig runs Rust tests using loaded config.
func (r *RustRunner) TestFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Test == nil {
		return fmt.Errorf("no test configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.TestProject(dir, r.config.Test.Args...)
}

// LintFromConfig runs Rust linting using loaded config.
func (r *RustRunner) LintFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Lint == nil {
		return fmt.Errorf("no lint configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.LintProject(dir, r.config.Lint.Tool, r.config.Lint.Args...)
}

// FormatFromConfig formats Rust code using loaded config.
func (r *RustRunner) FormatFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Format == nil {
		return fmt.Errorf("no format configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.FormatProject(dir, r.config.Format.Args...)
}

// CleanFromConfig cleans Rust project using loaded config.
func (r *RustRunner) CleanFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if r.config.Clean == nil {
		return fmt.Errorf("no clean configuration found")
	}

	dir := resolveDir(r.config.Directory)
	return r.CleanProject(dir, r.config.Clean.UseCargoClean, r.config.Clean.Paths...)
}

// RunGenCommandFromConfig runs a named command from genCommands configuration.
func (r *RustRunner) RunGenCommandFromConfig(commandName string) error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if len(r.config.GenCommands) == 0 {
		return fmt.Errorf("no genCommands configuration found")
	}

	var commandConfig *GenCommandConfig
	for _, command := range r.config.GenCommands {
		if command.Name == commandName {
			commandConfig = command
			break
		}
	}

	if commandConfig == nil {
		return fmt.Errorf("command '%s' not found in genCommands", commandName)
	}

	dir := resolveDir(r.config.Directory)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	if commandConfig.Command == "" {
		return fmt.Errorf("command '%s' has empty command", commandName)
	}

	return r.RunCargoCommand(dir, commandConfig.Command, commandConfig.Args...)
}

// RunAllGenCommandsFromConfig runs all commands from genCommands configuration in order.
func (r *RustRunner) RunAllGenCommandsFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if len(r.config.GenCommands) == 0 {
		return fmt.Errorf("no genCommands configuration found")
	}

	dir := resolveDir(r.config.Directory)
	if err := r.EnsureCargoProject(dir); err != nil {
		return err
	}

	slog.Info("🚀 Running all genCommands in order...", "count", len(r.config.GenCommands))

	for i, commandConfig := range r.config.GenCommands {
		if commandConfig.Command == "" {
			return fmt.Errorf("genCommand '%s' has empty command", commandConfig.Name)
		}

		slog.Info("Running command", "index", i+1, "name", commandConfig.Name, "command", commandConfig.Command)

		if err := r.RunCargoCommand(dir, commandConfig.Command, commandConfig.Args...); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", commandConfig.Name, err)
		}

		slog.Info("✅ Command completed", "name", commandConfig.Name)
	}

	slog.Info("✅ All genCommands completed successfully")
	return nil
}

// Package-level convenience functions for mage targets
var defaultRunner = NewRustRunner()

// LoadConfig loads Rust configuration from a YAML file.
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML.
func NewRunnerFromYAML(filepath string) (*RustRunner, error) {
	return NewRustRunnerFromYAML(filepath)
}

// Setup sets up Rust environment (requires loaded config).
func Setup() error {
	return defaultRunner.SetupFromConfig()
}

// Build builds Rust project (requires loaded config).
func Build() error {
	return defaultRunner.BuildFromConfig()
}

// Run runs Rust application (requires loaded config).
func Run() error {
	return defaultRunner.RunFromConfig()
}

// Test runs Rust tests (requires loaded config).
func Test() error {
	return defaultRunner.TestFromConfig()
}

// Lint runs Rust linting (requires loaded config).
func Lint() error {
	return defaultRunner.LintFromConfig()
}

// Format formats Rust code (requires loaded config).
func Format() error {
	return defaultRunner.FormatFromConfig()
}

// Clean cleans Rust project (requires loaded config).
func Clean() error {
	return defaultRunner.CleanFromConfig()
}

// RunAllGenCommands runs all commands from genCommands configuration in order.
func RunAllGenCommands() error {
	return defaultRunner.RunAllGenCommandsFromConfig()
}

// RunGenCommand runs a named command from genCommands configuration.
func RunGenCommand(commandName string) error {
	return defaultRunner.RunGenCommandFromConfig(commandName)
}
