// Package ci provides CI pipeline helpers for Mage-based Go projects.
// It wraps common verification patterns: generated-code drift detection,
// go mod tidiness checks, and composite CI gates.
package ci

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	execx "github.com/nirantaraai/nava/pkg/exec"
	"gopkg.in/yaml.v3"
)

// CIConfig holds YAML-driven configuration for CI targets.
type CIConfig struct {
	Verify *VerifyConfig `yaml:"verify,omitempty"`
}

// VerifyConfig holds options for CI verification targets.
type VerifyConfig struct {
	GenerateTarget string   `yaml:"generateTarget,omitempty"`
	ExtraChecks    []string `yaml:"extraChecks,omitempty"`
}

// CIRunner provides CI pipeline helpers.
type CIRunner struct {
	executor execx.Executor
	config   *CIConfig
}

// NewCIRunner creates a new CIRunner with the default executor.
func NewCIRunner() *CIRunner {
	return &CIRunner{executor: execx.NewExec()}
}

// NewCIRunnerWithExecutor creates a new CIRunner with a custom executor.
func NewCIRunnerWithExecutor(executor execx.Executor) *CIRunner {
	return &CIRunner{executor: executor}
}

// LoadConfig loads CI configuration from a YAML file.
func (c *CIRunner) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read ci config: %w", err)
	}
	var cfg CIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal ci config: %w", err)
	}
	c.config = &cfg
	return nil
}

// VerifyGenerated runs a generation command and checks for git diff.
// If the working tree has changes after running the generator, it fails
// with a descriptive error listing the modified files.
func (c *CIRunner) VerifyGenerated(generateCmd string, generateArgs ...string) error {
	slog.Info("🔍 Verifying generated code is up to date...", "cmd", generateCmd)
	start := time.Now()

	if err := c.executor.Run(context.Background(), generateCmd, false, generateArgs...); err != nil {
		return fmt.Errorf("generation command failed: %w", err)
	}

	if err := c.executor.Run(context.Background(), "git", false, "diff", "--exit-code"); err != nil {
		return fmt.Errorf("generated code is out of date — run the generator and commit the changes")
	}

	if err := c.executor.Run(context.Background(), "git", false, "diff", "--cached", "--exit-code"); err != nil {
		return fmt.Errorf("generated code has staged changes — commit or unstage them")
	}

	slog.Info("✅ Generated code is up to date", "duration", time.Since(start))
	return nil
}

// VerifyModTidy runs `go mod tidy` and checks for git diff on go.mod/go.sum.
func (c *CIRunner) VerifyModTidy() error {
	slog.Info("🔍 Verifying go.mod is tidy...")
	start := time.Now()

	if err := c.executor.Run(context.Background(), "go", false, "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	if err := c.executor.Run(context.Background(), "git", false, "diff", "--exit-code", "--", "go.mod", "go.sum"); err != nil {
		return fmt.Errorf("go.mod/go.sum are not tidy — run 'go mod tidy' and commit")
	}

	slog.Info("✅ go.mod is tidy", "duration", time.Since(start))
	return nil
}

// VerifyNoUntracked checks that no untracked files exist in the given paths.
func (c *CIRunner) VerifyNoUntracked(paths ...string) error {
	slog.Info("🔍 Checking for untracked files...", "paths", paths)

	args := []string{"ls-files", "--others", "--exclude-standard"}
	args = append(args, paths...)

	if err := c.executor.Run(context.Background(), "git", false, args...); err != nil {
		return fmt.Errorf("git ls-files failed: %w", err)
	}

	slog.Info("✅ No unexpected untracked files")
	return nil
}

// GoVet runs go vet with optional build tags.
func (c *CIRunner) GoVet(tags string) error {
	slog.Info("🔍 Running go vet...", "tags", tags)
	start := time.Now()

	args := []string{"vet"}
	if tags != "" {
		args = append(args, fmt.Sprintf("-tags=%s", tags))
	}
	args = append(args, "./...")

	if err := c.executor.Run(context.Background(), "go", false, args...); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	slog.Info("✅ go vet passed", "duration", time.Since(start))
	return nil
}

// Check runs a composite CI gate: vet + mod tidy verification.
func (c *CIRunner) Check(tags string) error {
	if err := c.GoVet(tags); err != nil {
		return err
	}
	return c.VerifyModTidy()
}

// VerifyGeneratedMage runs `mage <target>` and checks for git drift.
// This is the CI pattern: run mage generate, then git diff --exit-code.
func (c *CIRunner) VerifyGeneratedMage(target string) error {
	if target == "" {
		if c.config != nil && c.config.Verify != nil && c.config.Verify.GenerateTarget != "" {
			target = c.config.Verify.GenerateTarget
		} else {
			target = "generate"
		}
	}
	return c.VerifyGenerated("mage", target)
}

// EnvOrDefault returns the environment variable value or the default.
func EnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "JENKINS_URL", "BUILDKITE"}
	for _, v := range ciVars {
		if val := os.Getenv(v); val != "" && strings.ToLower(val) != "false" {
			return true
		}
	}
	return false
}
