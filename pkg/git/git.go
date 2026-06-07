package gitx

import (
	"context"
	"fmt"
	"strings"

	"github.com/nirantaraai/nava/pkg/exec"
)

// GitExecutor defines the interface for Git operations
type GitExecutor interface {
	GetCommitSHA() (string, error)
	GetShortCommitSHA() (string, error)
	GetBranch() (string, error)
	GetTag() (string, error)
	IsDirty() (bool, error)
	GetVersion() (string, error)
}

// GitRunner handles git command execution
type GitRunner struct {
	executor execx.Executor
}

// NewGitRunner creates a new GitRunner with the default executor
func NewGitRunner() *GitRunner {
	return &GitRunner{
		executor: execx.NewExec(),
	}
}

// NewGitRunnerWithExecutor creates a new GitRunner with a custom executor
func NewGitRunnerWithExecutor(executor execx.Executor) *GitRunner {
	return &GitRunner{
		executor: executor,
	}
}

// runGitCommand executes a git command and returns the output
func (g *GitRunner) runGitCommand(args ...string) (string, error) {
	creator := &execx.DefaultCommandCreator{}
	gitCmd := creator.CommandContext(context.Background(), "git", args...)

	output, err := gitCmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCommitSHA returns the current commit SHA
func (g *GitRunner) GetCommitSHA() (string, error) {
	output, err := g.runGitCommand("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}
	return output, nil
}

// GetShortCommitSHA returns the short commit SHA (7 characters)
func (g *GitRunner) GetShortCommitSHA() (string, error) {
	output, err := g.runGitCommand("rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get short commit SHA: %w", err)
	}
	return output, nil
}

// GetBranch returns the current branch name
func (g *GitRunner) GetBranch() (string, error) {
	output, err := g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}
	return output, nil
}

// GetTag returns the current tag if on a tag, empty string otherwise
func (g *GitRunner) GetTag() (string, error) {
	output, err := g.runGitCommand("describe", "--tags", "--exact-match")
	if err != nil {
		// Not on a tag, return empty string
		return "", nil
	}
	return output, nil
}

// IsDirty returns true if there are uncommitted changes
func (g *GitRunner) IsDirty() (bool, error) {
	output, err := g.runGitCommand("status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return len(output) > 0, nil
}

// GetVersion returns a version string based on git state
// Format: <tag>-<commits>-g<short-sha>[-dirty]
func (g *GitRunner) GetVersion() (string, error) {
	output, err := g.runGitCommand("describe", "--tags", "--always", "--dirty")
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	return output, nil
}

// Made with Bob
