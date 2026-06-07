package pythonmagex

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

// PythonRunner handles Python command execution with dependency injection
type PythonRunner struct {
	executor execx.Executor
	config   *PythonConfig
}

// NewPythonRunner creates a new PythonRunner with the default executor
func NewPythonRunner() *PythonRunner {
	return &PythonRunner{
		executor: execx.NewExec(),
	}
}

// NewPythonRunnerWithExecutor creates a new PythonRunner with a custom executor
func NewPythonRunnerWithExecutor(executor execx.Executor) *PythonRunner {
	return &PythonRunner{
		executor: executor,
	}
}

// NewPythonRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewPythonRunnerFromYAML(filepath string) (*PythonRunner, error) {
	runner := NewPythonRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads Python configuration from a YAML file
func (p *PythonRunner) LoadConfig(filepath string) error {
	config, err := LoadPythonConfig(filepath)
	if err != nil {
		return err
	}
	p.config = config
	return nil
}

// PythonConfig contains all Python operation configurations
type PythonConfig struct {
	Directory  string             `yaml:"directory,omitempty"`
	GenScripts []*GenScriptConfig `yaml:"genScripts,omitempty"`
	Setup      *SetupOptions      `yaml:"setup,omitempty"`
	RunScript  *RunScriptOptions  `yaml:"runScript,omitempty"`
	RunService *RunScriptOptions  `yaml:"runService,omitempty"`
	RunTests   *RunTestsOptions   `yaml:"runTests,omitempty"`
	RunLint    *RunLintOptions    `yaml:"runLint,omitempty"`
	RunFormat  *RunFormatOptions  `yaml:"runFormat,omitempty"`
	Clean      *CleanConfig       `yaml:"clean,omitempty"`
}

// GenScriptConfig contains configuration for a named script
type GenScriptConfig struct {
	Name   string   `yaml:"name"`
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// SetupOptions contains options for setting up Python environment
type SetupOptions struct {
	CreateVenv          bool `yaml:"createVenv,omitempty"`
	InstallRequirements bool `yaml:"installRequirements,omitempty"`
}

// RunScriptOptions contains options for running Python scripts
type RunScriptOptions struct {
	Script string   `yaml:"script"`
	Args   []string `yaml:"args,omitempty"`
}

// RunTestsOptions contains options for running Python tests
type RunTestsOptions struct {
	Args []string `yaml:"args,omitempty"`
}

// RunLintOptions contains options for running Python linting
type RunLintOptions struct {
	Linter string   `yaml:"linter,omitempty"`
	Args   []string `yaml:"args,omitempty"`
}

// RunFormatOptions contains options for formatting Python code
type RunFormatOptions struct {
	Args []string `yaml:"args,omitempty"`
}

// CleanConfig contains options for cleaning Python project
type CleanConfig struct {
	Paths []string `yaml:"paths,omitempty"`
}

// LoadPythonConfig loads Python configuration from a YAML file
func LoadPythonConfig(filepath string) (*PythonConfig, error) {
	var config PythonConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// CreateVenv creates a Python virtual environment
func (p *PythonRunner) CreateVenv(dir string) error {
	venvPath := filepath.Join(dir, "venv")

	if _, err := os.Stat(venvPath); err == nil {
		slog.Info("✅ Virtual environment already exists", "path", venvPath)
		return nil
	}

	slog.Info("📦 Creating Python virtual environment...", "path", venvPath)
	start := time.Now()

	if err := p.executor.Run(context.Background(), "python3", false, "-m", "venv", venvPath); err != nil {
		return fmt.Errorf("failed to create venv: %w", err)
	}

	slog.Info("✅ Virtual environment created", "duration", time.Since(start))
	return nil
}

// InstallRequirements installs Python packages from requirements.txt
func (p *PythonRunner) InstallRequirements(dir string) error {
	reqFile := filepath.Join(dir, "requirements.txt")

	if _, err := os.Stat(reqFile); err != nil {
		return fmt.Errorf("requirements.txt not found in %s", dir)
	}

	pipPath := filepath.Join(dir, "venv", "bin", "pip")

	slog.Info("📦 Installing Python dependencies...", "file", reqFile)
	start := time.Now()

	if err := p.executor.Run(context.Background(), pipPath, false, "install", "-r", reqFile); err != nil {
		return fmt.Errorf("failed to install requirements: %w", err)
	}

	slog.Info("✅ Dependencies installed", "duration", time.Since(start))
	return nil
}

// RunScript runs a Python script using the virtual environment
func (p *PythonRunner) RunScript(dir, script string, args ...string) error {
	pythonPath := filepath.Join(dir, "venv", "bin", "python")
	scriptPath := filepath.Join(dir, script)

	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	slog.Info("🐍 Running Python script...", "script", script)
	start := time.Now()

	cmdArgs := append([]string{scriptPath}, args...)
	if err := p.executor.Run(context.Background(), pythonPath, false, cmdArgs...); err != nil {
		return fmt.Errorf("failed to run script: %w", err)
	}

	slog.Info("✅ Script completed", "duration", time.Since(start))
	return nil
}

// RunTests runs Python tests using pytest
func (p *PythonRunner) RunTests(dir string, args ...string) error {
	pythonPath := filepath.Join(dir, "venv", "bin", "python")

	slog.Info("🧪 Running Python tests...")
	start := time.Now()

	defaultArgs := []string{"-m", "pytest"}
	cmdArgs := append(defaultArgs, args...)

	if err := p.executor.Run(context.Background(), pythonPath, false, cmdArgs...); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	slog.Info("✅ Tests passed", "duration", time.Since(start))
	return nil
}

// RunLint runs Python linting using flake8 or pylint
func (p *PythonRunner) RunLint(dir string, linter string, args ...string) error {
	pythonPath := filepath.Join(dir, "venv", "bin", "python")

	if linter == "" {
		linter = "flake8"
	}

	slog.Info("🔍 Running Python linter...", "linter", linter)
	start := time.Now()

	cmdArgs := append([]string{"-m", linter}, args...)
	if err := p.executor.Run(context.Background(), pythonPath, false, cmdArgs...); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	slog.Info("✅ Linting passed", "duration", time.Since(start))
	return nil
}

// RunFormat formats Python code using black
func (p *PythonRunner) RunFormat(dir string, args ...string) error {
	pythonPath := filepath.Join(dir, "venv", "bin", "python")

	slog.Info("✨ Formatting Python code...")
	start := time.Now()

	defaultArgs := []string{"-m", "black", "."}
	cmdArgs := append(defaultArgs, args...)

	if err := p.executor.Run(context.Background(), pythonPath, false, cmdArgs...); err != nil {
		return fmt.Errorf("formatting failed: %w", err)
	}

	slog.Info("✅ Formatting complete", "duration", time.Since(start))
	return nil
}

// Setup sets up a complete Python environment (venv + requirements)
func (p *PythonRunner) Setup(dir string) error {
	slog.Info("🎯 Setting up Python environment...", "directory", dir)

	if err := p.CreateVenv(dir); err != nil {
		return err
	}

	if err := p.InstallRequirements(dir); err != nil {
		return err
	}

	slog.Info("✅ Python environment setup complete")
	return nil
}

// SetupFromConfig sets up Python environment using loaded config
func (p *PythonRunner) SetupFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.Setup == nil {
		return fmt.Errorf("no setup configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🎯 Setting up Python environment from config...", "directory", dir)

	if p.config.Setup.CreateVenv {
		if err := p.CreateVenv(dir); err != nil {
			return err
		}
	}

	if p.config.Setup.InstallRequirements {
		if err := p.InstallRequirements(dir); err != nil {
			return err
		}
	}

	slog.Info("✅ Python environment setup complete")
	return nil
}

// RunScriptFromConfig runs a Python script using loaded config
func (p *PythonRunner) RunScriptFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.RunScript == nil {
		return fmt.Errorf("no runScript configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	return p.RunScript(dir, p.config.RunScript.Script, p.config.RunScript.Args...)
}

// RunServiceFromConfig runs a Python service using loaded config
func (p *PythonRunner) RunServiceFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.RunService == nil {
		return fmt.Errorf("no runService configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	return p.RunScript(dir, p.config.RunService.Script, p.config.RunService.Args...)
}

// RunTestsFromConfig runs Python tests using loaded config
func (p *PythonRunner) RunTestsFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.RunTests == nil {
		return fmt.Errorf("no runTests configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	return p.RunTests(dir, p.config.RunTests.Args...)
}

// RunLintFromConfig runs Python linting using loaded config
func (p *PythonRunner) RunLintFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.RunLint == nil {
		return fmt.Errorf("no runLint configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	linter := p.config.RunLint.Linter
	if linter == "" {
		linter = "flake8"
	}

	return p.RunLint(dir, linter, p.config.RunLint.Args...)
}

// RunFormatFromConfig formats Python code using loaded config
func (p *PythonRunner) RunFormatFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.RunFormat == nil {
		return fmt.Errorf("no runFormat configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	return p.RunFormat(dir, p.config.RunFormat.Args...)
}

// CleanFromConfig cleans Python project using loaded config
func (p *PythonRunner) CleanFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.Clean == nil {
		return fmt.Errorf("no clean configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🧹 Cleaning Python project...", "directory", dir)

	defaultPaths := []string{"venv", "__pycache__", "*.pyc", ".pytest_cache", "models"}
	paths := p.config.Clean.Paths
	if len(paths) == 0 {
		paths = defaultPaths
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

// RunGenScriptFromConfig runs a named script from genScripts configuration
func (p *PythonRunner) RunGenScriptFromConfig(scriptName string) error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.GenScripts == nil || len(p.config.GenScripts) == 0 {
		return fmt.Errorf("no genScripts configuration found")
	}

	// Find the script by name
	var scriptConfig *GenScriptConfig
	for _, script := range p.config.GenScripts {
		if script.Name == scriptName {
			scriptConfig = script
			break
		}
	}

	if scriptConfig == nil {
		return fmt.Errorf("script '%s' not found in genScripts", scriptName)
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	return p.RunScript(dir, scriptConfig.Script, scriptConfig.Args...)
}

// RunAllGenScriptsFromConfig runs all scripts from genScripts configuration in order
func (p *PythonRunner) RunAllGenScriptsFromConfig() error {
	if p.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if p.config.GenScripts == nil || len(p.config.GenScripts) == 0 {
		return fmt.Errorf("no genScripts configuration found")
	}

	dir := p.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("🚀 Running all genScripts in order...", "count", len(p.config.GenScripts))

	for i, scriptConfig := range p.config.GenScripts {
		slog.Info("Running script", "index", i+1, "name", scriptConfig.Name, "script", scriptConfig.Script)

		if err := p.RunScript(dir, scriptConfig.Script, scriptConfig.Args...); err != nil {
			return fmt.Errorf("failed to run script '%s': %w", scriptConfig.Name, err)
		}

		slog.Info("✅ Script completed", "name", scriptConfig.Name)
	}

	slog.Info("✅ All genScripts completed successfully")
	return nil
}

// Package-level convenience functions for mage targets
var defaultRunner = NewPythonRunner()

// LoadConfig loads Python configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*PythonRunner, error) {
	return NewPythonRunnerFromYAML(filepath)
}

// Setup sets up Python environment (requires loaded config)
func Setup() error {
	return defaultRunner.SetupFromConfig()
}

// RunAllGenScripts runs all scripts from genScripts configuration in order
func RunAllGenScripts() error {
	return defaultRunner.RunAllGenScriptsFromConfig()
}

// RunScript runs a Python script (requires loaded config)
func RunScript() error {
	return defaultRunner.RunScriptFromConfig()
}

// RunGenScript runs a named script from genScripts configuration
func RunGenScript(scriptName string) error {
	return defaultRunner.RunGenScriptFromConfig(scriptName)

}

// RunService runs a Python service (requires loaded config)
func RunService() error {
	return defaultRunner.RunServiceFromConfig()
}

// RunTests runs Python tests (requires loaded config)
func RunTests() error {
	return defaultRunner.RunTestsFromConfig()
}

// RunLint runs Python linting (requires loaded config)
func RunLint() error {
	return defaultRunner.RunLintFromConfig()
}

// RunFormat formats Python code (requires loaded config)
func RunFormat() error {
	return defaultRunner.RunFormatFromConfig()
}

// Clean cleans Python project (requires loaded config)
func Clean() error {
	return defaultRunner.CleanFromConfig()
}
