package goreleaserx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// GoReleaserRunner handles goreleaser command execution with dependency injection
type GoReleaserRunner struct {
	executor execx.Executor
	config   *GoReleaserConfig
}

// GoReleaserConfig contains goreleaser operation configuration
type GoReleaserConfig struct {
	Config   string   `yaml:"config,omitempty"`   // path to .goreleaser.yaml
	Clean    bool     `yaml:"clean,omitempty"`    // --clean: remove dist/ before build
	Snapshot bool     `yaml:"snapshot,omitempty"` // --snapshot: non-publishing local release
	Args     []string `yaml:"args,omitempty"`     // extra goreleaser arguments
}

// NewGoReleaserRunner creates a new GoReleaserRunner with the default executor
func NewGoReleaserRunner() *GoReleaserRunner {
	return &GoReleaserRunner{
		executor: execx.NewExec(),
	}
}

// NewGoReleaserRunnerWithExecutor creates a new GoReleaserRunner with a custom executor
func NewGoReleaserRunnerWithExecutor(executor execx.Executor) *GoReleaserRunner {
	return &GoReleaserRunner{executor: executor}
}

// NewGoReleaserRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewGoReleaserRunnerFromYAML(path string) (*GoReleaserRunner, error) {
	config, err := LoadGoReleaserConfig(path)
	if err != nil {
		return nil, err
	}
	return &GoReleaserRunner{
		executor: execx.NewExec(),
		config:   config,
	}, nil
}

// LoadConfig loads goreleaser configuration from a YAML file
func (r *GoReleaserRunner) LoadConfig(path string) error {
	config, err := LoadGoReleaserConfig(path)
	if err != nil {
		return err
	}
	r.config = config
	return nil
}

// LoadGoReleaserConfig loads goreleaser configuration from a YAML file
func LoadGoReleaserConfig(path string) (*GoReleaserConfig, error) {
	var config GoReleaserConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return &config, nil
}

// releaseArgs builds the goreleaser argument list for a release subcommand
func (r *GoReleaserRunner) releaseArgs(forceSnapshot bool) []string {
	args := []string{"release"}
	if r.config.Clean {
		args = append(args, "--clean")
	}
	if r.config.Config != "" {
		args = append(args, "--config", r.config.Config)
	}
	if forceSnapshot || r.config.Snapshot {
		args = append(args, "--snapshot")
	}
	return append(args, r.config.Args...)
}

// ReleaseFromConfig runs goreleaser release using loaded config
func (r *GoReleaserRunner) ReleaseFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no goreleaser configuration loaded")
	}

	slog.Info("Running goreleaser release...")
	start := time.Now()

	if err := r.executor.Run(context.Background(), "goreleaser", false, r.releaseArgs(false)...); err != nil {
		return err
	}

	slog.Info("goreleaser release complete", "duration", time.Since(start))
	return nil
}

// SnapshotFromConfig runs a local snapshot release (no publish) using loaded config
func (r *GoReleaserRunner) SnapshotFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no goreleaser configuration loaded")
	}

	slog.Info("Running goreleaser snapshot...")
	start := time.Now()

	if err := r.executor.Run(context.Background(), "goreleaser", false, r.releaseArgs(true)...); err != nil {
		return err
	}

	slog.Info("goreleaser snapshot complete", "duration", time.Since(start))
	return nil
}

// CheckFromConfig validates the goreleaser configuration file
func (r *GoReleaserRunner) CheckFromConfig() error {
	if r.config == nil {
		return fmt.Errorf("no goreleaser configuration loaded")
	}

	args := []string{"check"}
	if r.config.Config != "" {
		args = append(args, "--config", r.config.Config)
	}
	return r.executor.Run(context.Background(), "goreleaser", false, args...)
}
