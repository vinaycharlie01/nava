package nodejsmagex

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

// NodeRunner handles Node.js command execution with dependency injection
type NodeRunner struct {
	executor execx.Executor
	config   *NodeConfig
}

// NewNodeRunner creates a new NodeRunner with the default executor
func NewNodeRunner() *NodeRunner {
	return &NodeRunner{
		executor: execx.NewExec(),
	}
}

// NewNodeRunnerWithExecutor creates a new NodeRunner with a custom executor
func NewNodeRunnerWithExecutor(executor execx.Executor) *NodeRunner {
	return &NodeRunner{
		executor: executor,
	}
}

// NewNodeRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewNodeRunnerFromYAML(filepath string) (*NodeRunner, error) {
	runner := NewNodeRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads Node.js configuration from a YAML file
func (n *NodeRunner) LoadConfig(filepath string) error {
	config, err := LoadNodeConfig(filepath)
	if err != nil {
		return err
	}
	n.config = config
	return nil
}

// NodeConfig contains all Node.js operation configurations
type NodeConfig struct {
	Directory string           `yaml:"directory,omitempty"`
	Setup     *SetupOptions    `yaml:"setup,omitempty"`
	Run       *RunScriptConfig `yaml:"run,omitempty"`
	Build     *BuildConfig     `yaml:"build,omitempty"`
	Test      *TestConfig      `yaml:"test,omitempty"`
	Lint      *LintConfig      `yaml:"lint,omitempty"`
	Format    *FormatConfig    `yaml:"format,omitempty"`
	Clean     *CleanConfig     `yaml:"clean,omitempty"`
	Preview   *PreviewConfig   `yaml:"preview,omitempty"`
	Dev       *DevConfig       `yaml:"dev,omitempty"`
}

// SetupOptions contains options for setting up Node.js environment
type SetupOptions struct {
	Install bool `yaml:"install,omitempty"`
}

// RunScriptConfig contains options for running npm scripts
type RunScriptConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// BuildConfig contains options for building Node.js projects
type BuildConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// TestConfig contains options for running Node.js tests
type TestConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// LintConfig contains options for running linting
type LintConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// FormatConfig contains options for formatting code
type FormatConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// CleanConfig contains options for cleaning build artifacts
type CleanConfig struct {
	Paths []string `yaml:"paths,omitempty"`
}

// PreviewConfig contains options for running preview server
type PreviewConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// DevConfig contains options for running development server
type DevConfig struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// LoadNodeConfig loads Node.js configuration from a YAML file
func LoadNodeConfig(filepath string) (*NodeConfig, error) {
	var config NodeConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// InstallDependencies installs Node.js dependencies using npm
func (n *NodeRunner) InstallDependencies(dir string) error {
	packageJSON := filepath.Join(dir, "package.json")

	if _, err := os.Stat(packageJSON); err != nil {
		return fmt.Errorf("package.json not found in %s", dir)
	}

	slog.Info("📦 Installing Node.js dependencies...", "directory", dir)
	start := time.Now()

	if err := n.executor.RunInDir(context.Background(), dir, "npm", false, "install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	slog.Info("✅ Dependencies installed", "duration", time.Since(start))
	return nil
}

// RunScript runs an npm script defined in package.json
func (n *NodeRunner) RunScript(dir, script string, args ...string) error {
	slog.Info("🚀 Running npm script...", "script", script, "directory", dir)
	start := time.Now()

	cmdArgs := append([]string{"run", script}, args...)
	if err := n.executor.RunInDir(context.Background(), dir, "npm", false, cmdArgs...); err != nil {
		return fmt.Errorf("failed to run script: %w", err)
	}

	slog.Info("✅ Script completed", "duration", time.Since(start))
	return nil
}

// SetupFromConfig sets up Node.js environment using loaded config
func (n *NodeRunner) SetupFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Setup == nil {
		return fmt.Errorf("no setup configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🎯 Setting up Node.js environment from config...", "directory", dir)

	if n.config.Setup.Install {
		if err := n.InstallDependencies(dir); err != nil {
			return err
		}
	}

	slog.Info("✅ Node.js environment setup complete")
	return nil
}

// RunFromConfig runs npm script using loaded config
func (n *NodeRunner) RunFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Run == nil {
		return fmt.Errorf("no run configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Run.Script, n.config.Run.Args...)
}

// BuildFromConfig builds Node.js project using loaded config
func (n *NodeRunner) BuildFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Build == nil {
		return fmt.Errorf("no build configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Build.Script, n.config.Build.Args...)
}

// TestFromConfig runs Node.js tests using loaded config
func (n *NodeRunner) TestFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Test == nil {
		return fmt.Errorf("no test configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Test.Script, n.config.Test.Args...)
}

// DevFromConfig runs development server using loaded config
func (n *NodeRunner) DevFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Dev == nil {
		return fmt.Errorf("no dev configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Dev.Script, n.config.Dev.Args...)
}

// PreviewFromConfig runs preview server using loaded config
func (n *NodeRunner) PreviewFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Preview == nil {
		return fmt.Errorf("no preview configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Preview.Script, n.config.Preview.Args...)
}

// LintFromConfig runs linting using loaded config
func (n *NodeRunner) LintFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Lint == nil {
		return fmt.Errorf("no lint configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Lint.Script, n.config.Lint.Args...)
}

// FormatFromConfig formats code using loaded config
func (n *NodeRunner) FormatFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Format == nil {
		return fmt.Errorf("no format configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	return n.RunScript(dir, n.config.Format.Script, n.config.Format.Args...)
}

// CleanFromConfig cleans build artifacts using loaded config
func (n *NodeRunner) CleanFromConfig() error {
	if n.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if n.config.Clean == nil {
		return fmt.Errorf("no clean configuration found")
	}

	dir := n.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🧹 Cleaning Node.js project...", "directory", dir)

	defaultPaths := []string{"node_modules", "dist", "build"}
	paths := n.config.Clean.Paths
	if len(paths) == 0 {
		paths = defaultPaths
	}

	for _, path := range paths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			slog.Info("Removing directory", "path", fullPath)
			if err := os.RemoveAll(fullPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", fullPath, err)
			}
		}
	}

	slog.Info("✅ Clean complete")
	return nil
}

// Package-level convenience functions for mage targets
var defaultRunner = NewNodeRunner()

// LoadConfig loads Node.js configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*NodeRunner, error) {
	return NewNodeRunnerFromYAML(filepath)
}

// Setup sets up Node.js environment (requires loaded config)
func Setup() error {
	return defaultRunner.SetupFromConfig()
}

// Run runs npm script (requires loaded config)
func Run() error {
	return defaultRunner.RunFromConfig()
}

// Build builds Node.js project (requires loaded config)
func Build() error {
	return defaultRunner.BuildFromConfig()
}

// Test runs Node.js tests (requires loaded config)
func Test() error {
	return defaultRunner.TestFromConfig()
}

// Dev runs development server (requires loaded config)
func Dev() error {
	return defaultRunner.DevFromConfig()
}

// Preview runs preview server (requires loaded config)
func Preview() error {
	return defaultRunner.PreviewFromConfig()
}

// Lint runs linting (requires loaded config)
func Lint() error {
	return defaultRunner.LintFromConfig()
}

// Format formats code (requires loaded config)
func Format() error {
	return defaultRunner.FormatFromConfig()
}

// Clean cleans build artifacts (requires loaded config)
func Clean() error {
	return defaultRunner.CleanFromConfig()
}
