package golangmagex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	gitx "github.com/nirantaraai/nava/pkg/git"
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
	Directory   string             `yaml:"directory,omitempty"`
	Setup       *SetupOptions      `yaml:"setup,omitempty"`
	Build       *BuildConfig       `yaml:"build,omitempty"`
	Run         *RunConfig         `yaml:"run,omitempty"`
	Test        *TestConfig        `yaml:"test,omitempty"`
	Race        *RaceConfig        `yaml:"race,omitempty"`
	Coverage    *CoverageConfig    `yaml:"coverage,omitempty"`
	Bench       *BenchConfig       `yaml:"bench,omitempty"`
	Lint        *LintConfig        `yaml:"lint,omitempty"`
	Vet         *VetConfig         `yaml:"vet,omitempty"`
	Format      *FormatConfig      `yaml:"format,omitempty"`
	Install     *InstallConfig     `yaml:"install,omitempty"`
	Govulncheck *GovulncheckConfig `yaml:"govulncheck,omitempty"`
	CrossBuild  *CrossBuildConfig  `yaml:"crossBuild,omitempty"`
}

// SetupOptions contains options for setting up Go environment
type SetupOptions struct {
	ModDownload bool `yaml:"modDownload,omitempty"`
	ModTidy     bool `yaml:"modTidy,omitempty"`
}

// BuildConfig contains options for building Go binaries
type BuildConfig struct {
	Output     string   `yaml:"output"`
	Main       string   `yaml:"main"`
	Args       []string `yaml:"args,omitempty"`
	VersionPkg string   `yaml:"versionPkg,omitempty"` // auto-inject Version/Commit/BuildDate ldflags via git
	LDFlags    string   `yaml:"ldflags,omitempty"`    // raw ldflags string (overrides versionPkg)
	NoCGO      bool     `yaml:"noCGO,omitempty"`      // set CGO_ENABLED=0 for a static binary
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

// RaceConfig contains options for running Go race tests
type RaceConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Args     []string `yaml:"args,omitempty"`
}

// CoverageConfig contains options for running Go test coverage
type CoverageConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Args     []string `yaml:"args,omitempty"`
	Output   string   `yaml:"output,omitempty"`
}

// BenchConfig contains options for running Go benchmarks
type BenchConfig struct {
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

// GovulncheckConfig contains options for running govulncheck
type GovulncheckConfig struct {
	Packages []string `yaml:"packages,omitempty"`
}

// CrossBuildConfig contains options for cross-compiling Go binaries
type CrossBuildConfig struct {
	Main       string           `yaml:"main"`
	Binary     string           `yaml:"binary"`
	OutputDir  string           `yaml:"outputDir,omitempty"`
	VersionPkg string           `yaml:"versionPkg,omitempty"`
	LDFlags    string           `yaml:"ldflags,omitempty"`
	Platforms  []PlatformTarget `yaml:"platforms"`
}

// PlatformTarget specifies an OS/arch combination to cross-compile for
type PlatformTarget struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
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
	slog.Info("Running Go command...", "dir", dir, "command", command)
	start := time.Now()

	cmdArgs := append([]string{command}, args...)
	if err := g.executor.RunInDir(context.Background(), dir, "go", false, cmdArgs...); err != nil {
		return err
	}

	slog.Info("Command completed", "duration", time.Since(start))
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

	slog.Info("Setting up Go environment...", "directory", dir)

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

	slog.Info("Go environment setup complete")
	return nil
}

// BuildFromConfig builds Go binary using loaded config.
// When versionPkg is set, Version/Commit/BuildDate ldflags are auto-injected via git.
// When noCGO is true, CGO_ENABLED=0 is set for a static binary.
func (g *GoRunner) BuildFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Build == nil {
		return fmt.Errorf("no build configuration found")
	}

	cfg := g.config.Build
	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("Building Go binary...", "directory", dir, "output", cfg.Output)
	start := time.Now()

	// Resolve ldflags: explicit value wins; versionPkg triggers git injection
	ldflags := cfg.LDFlags
	if ldflags == "" && cfg.VersionPkg != "" {
		git := gitx.NewGitRunner()
		version, _ := git.GetVersion()
		commit, _ := git.GetShortCommitSHA()
		date := time.Now().UTC().Format(time.RFC3339)
		ldflags = fmt.Sprintf("-s -w -X %s.Version=%s -X %s.Commit=%s -X %s.BuildDate=%s",
			cfg.VersionPkg, version,
			cfg.VersionPkg, commit,
			cfg.VersionPkg, date)
	}

	if ldflags != "" || cfg.NoCGO {
		// Need env control: use exec.Command directly
		buildArgs := []string{"build"}
		if ldflags != "" {
			buildArgs = append(buildArgs, "-ldflags", ldflags)
		}
		buildArgs = append(buildArgs, cfg.Args...)
		buildArgs = append(buildArgs, "-o", cfg.Output, cfg.Main)

		if outDir := filepath.Dir(cfg.Output); outDir != "." {
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}
		}

		cmd := exec.Command("go", buildArgs...) //nolint:gosec
		if cfg.NoCGO {
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		}
		if dir != "." {
			cmd.Dir = dir
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build: %w", err)
		}
	} else {
		buildArgs := append([]string{"-o", cfg.Output, cfg.Main}, cfg.Args...)
		if err := g.RunInDir(dir, "build", buildArgs...); err != nil {
			return err
		}
	}

	slog.Info("Build complete", "duration", time.Since(start))
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	slog.Info("Running Go program...", "directory", dir)

	runArgs := append([]string{g.config.Run.Main}, g.config.Run.Args...)
	if err := g.executor.RunInDir(ctx, dir, "go", false, append([]string{"run"}, runArgs...)...); err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("Program stopped gracefully")
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

	slog.Info("Running Go tests...", "directory", dir)
	start := time.Now()

	testArgs := append(packages, g.config.Test.Args...)
	if err := g.RunInDir(dir, "test", testArgs...); err != nil {
		return err
	}

	slog.Info("Tests passed", "duration", time.Since(start))
	return nil
}

// RaceFromConfig runs Go tests with race detection using loaded config
func (g *GoRunner) RaceFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Race == nil {
		return fmt.Errorf("no race configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := g.config.Race.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	slog.Info("Running race tests...", "directory", dir)
	start := time.Now()

	raceArgs := append([]string{"-race"}, packages...)
	raceArgs = append(raceArgs, g.config.Race.Args...)
	if err := g.RunInDir(dir, "test", raceArgs...); err != nil {
		return err
	}

	slog.Info("Race tests passed", "duration", time.Since(start))
	return nil
}

// CoverageFromConfig runs Go tests with coverage profiling using loaded config
func (g *GoRunner) CoverageFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Coverage == nil {
		return fmt.Errorf("no coverage configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := g.config.Coverage.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	output := g.config.Coverage.Output
	if output == "" {
		output = "coverage.out"
	}

	slog.Info("Running coverage...", "directory", dir)
	start := time.Now()

	coverArgs := append([]string{"-coverprofile=" + output, "-covermode=atomic"}, packages...)
	coverArgs = append(coverArgs, g.config.Coverage.Args...)
	if err := g.RunInDir(dir, "test", coverArgs...); err != nil {
		return err
	}

	slog.Info("Coverage complete", "duration", time.Since(start), "output", output)
	return nil
}

// BenchFromConfig runs Go benchmarks using loaded config
func (g *GoRunner) BenchFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	if g.config.Bench == nil {
		return fmt.Errorf("no bench configuration found")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := g.config.Bench.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	slog.Info("Running benchmarks...", "directory", dir)
	start := time.Now()

	benchArgs := append([]string{"-bench=.", "-benchmem", "-run=^$"}, packages...)
	benchArgs = append(benchArgs, g.config.Bench.Args...)
	if err := g.RunInDir(dir, "test", benchArgs...); err != nil {
		return err
	}

	slog.Info("Benchmarks complete", "duration", time.Since(start))
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

	slog.Info("Running golangci-lint...", "directory", dir)
	start := time.Now()

	lintArgs := append([]string{"run", "--timeout=5m"}, g.config.Lint.Args...)
	if err := g.executor.RunInDir(context.Background(), dir, "golangci-lint", false, lintArgs...); err != nil {
		return err
	}

	slog.Info("Linting passed", "duration", time.Since(start))
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

	slog.Info("Running go vet...", "directory", dir)
	start := time.Now()

	vetArgs := append(packages, g.config.Vet.Args...)
	if err := g.RunInDir(dir, "vet", vetArgs...); err != nil {
		return err
	}

	slog.Info("Go vet passed", "duration", time.Since(start))
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

	slog.Info("Formatting Go code...", "directory", dir)
	start := time.Now()

	formatArgs := append([]string{"-w", "."}, g.config.Format.Args...)
	if err := g.executor.RunInDir(context.Background(), dir, "gofmt", false, formatArgs...); err != nil {
		return err
	}

	slog.Info("Formatting complete", "duration", time.Since(start))
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
		slog.Info("No packages to install")
		return nil
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	slog.Info("Installing Go packages...", "directory", dir, "packages", g.config.Install.Packages)
	start := time.Now()

	for _, pkg := range g.config.Install.Packages {
		installArgs := append([]string{pkg}, g.config.Install.Args...)
		if err := g.RunInDir(dir, "install", installArgs...); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}

	slog.Info("Installation complete", "duration", time.Since(start))
	return nil
}

// GovulncheckFromConfig runs govulncheck using loaded config
func (g *GoRunner) GovulncheckFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	dir := g.config.Directory
	if dir == "" {
		dir = "."
	}

	packages := []string{"./..."}
	if g.config.Govulncheck != nil && len(g.config.Govulncheck.Packages) > 0 {
		packages = g.config.Govulncheck.Packages
	}

	slog.Info("Running govulncheck...", "directory", dir)
	start := time.Now()

	if err := g.executor.RunInDir(context.Background(), dir, "govulncheck", false, packages...); err != nil {
		return err
	}

	slog.Info("Govulncheck passed", "duration", time.Since(start))
	return nil
}

// CrossBuildFromConfig cross-compiles Go binaries for each configured platform.
// When versionPkg is set, Version/Commit/BuildDate ldflags are auto-injected via git.
func (g *GoRunner) CrossBuildFromConfig() error {
	if g.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	cfg := g.config.CrossBuild
	if cfg == nil {
		return fmt.Errorf("no crossBuild configuration found")
	}
	if len(cfg.Platforms) == 0 {
		return fmt.Errorf("crossBuild.platforms must not be empty")
	}

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = "dist"
	}

	// Resolve ldflags: explicit value takes precedence; versionPkg triggers git injection
	ldflags := cfg.LDFlags
	if ldflags == "" && cfg.VersionPkg != "" {
		git := gitx.NewGitRunner()
		version, _ := git.GetVersion()
		commit, _ := git.GetShortCommitSHA()
		date := time.Now().UTC().Format(time.RFC3339)
		ldflags = fmt.Sprintf("-s -w -X %s.Version=%s -X %s.Commit=%s -X %s.BuildDate=%s",
			cfg.VersionPkg, version,
			cfg.VersionPkg, commit,
			cfg.VersionPkg, date)
	}

	slog.Info("Cross-compiling...", "platforms", len(cfg.Platforms), "binary", cfg.Binary)
	start := time.Now()

	for _, p := range cfg.Platforms {
		outDir := filepath.Join(outputDir, p.OS+"_"+p.Arch)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		out := filepath.Join(outDir, cfg.Binary)

		slog.Info("Cross-compiling platform", "os", p.OS, "arch", p.Arch, "output", out)

		buildArgs := []string{"build"}
		if ldflags != "" {
			buildArgs = append(buildArgs, "-ldflags", ldflags)
		}
		buildArgs = append(buildArgs, "-o", out, cfg.Main)

		cmd := exec.Command("go", buildArgs...) //nolint:gosec
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+p.OS, "GOARCH="+p.Arch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cross-build %s/%s: %w", p.OS, p.Arch, err)
		}
	}

	slog.Info("Cross-build complete", "duration", time.Since(start))
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
func Setup() error { return defaultRunner.SetupFromConfig() }

// Build builds Go binary (requires loaded config)
func Build() error { return defaultRunner.BuildFromConfig() }

// Run runs Go program (requires loaded config)
func Run() error { return defaultRunner.RunFromConfig() }

// Test runs Go tests (requires loaded config)
func Test() error { return defaultRunner.TestFromConfig() }

// Race runs tests with race detection (requires loaded config)
func Race() error { return defaultRunner.RaceFromConfig() }

// Coverage runs tests with coverage profiling (requires loaded config)
func Coverage() error { return defaultRunner.CoverageFromConfig() }

// Bench runs benchmarks (requires loaded config)
func Bench() error { return defaultRunner.BenchFromConfig() }

// Lint runs golangci-lint (requires loaded config)
func Lint() error { return defaultRunner.LintFromConfig() }

// Vet runs go vet (requires loaded config)
func Vet() error { return defaultRunner.VetFromConfig() }

// Format formats Go code (requires loaded config)
func Format() error { return defaultRunner.FormatFromConfig() }

// Install installs Go packages (requires loaded config)
func Install() error { return defaultRunner.InstallFromConfig() }

// Govulncheck runs govulncheck (requires loaded config)
func Govulncheck() error { return defaultRunner.GovulncheckFromConfig() }

// CrossBuild cross-compiles for all configured platforms (requires loaded config)
func CrossBuild() error { return defaultRunner.CrossBuildFromConfig() }
