package gitmagex

import (
	"github.com/nirantaraai/nava/pkg/git"
)

// Package-level convenience functions for mage targets
var defaultRunner = gitx.NewGitRunner()

// GetCommitSHA returns the current commit SHA
func GetCommitSHA() (string, error) {
	return defaultRunner.GetCommitSHA()
}

// GetShortCommitSHA returns the short commit SHA (7 characters)
func GetShortCommitSHA() (string, error) {
	return defaultRunner.GetShortCommitSHA()
}

// GetBranch returns the current branch name
func GetBranch() (string, error) {
	return defaultRunner.GetBranch()
}

// GetTag returns the current tag if on a tag, empty string otherwise
func GetTag() (string, error) {
	return defaultRunner.GetTag()
}

// IsDirty returns true if there are uncommitted changes
func IsDirty() (bool, error) {
	return defaultRunner.IsDirty()
}

// GetVersion returns a version string based on git state
func GetVersion() (string, error) {
	return defaultRunner.GetVersion()
}

// Made with Bob
