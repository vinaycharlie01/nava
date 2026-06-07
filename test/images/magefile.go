//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/nirantaraai/nava/mage/docker"
	"github.com/nirantaraai/nava/mage/git"
)

// Docker namespace for Docker operations
type Docker mg.Namespace

// Git namespace for Git operations
type Git mg.Namespace

// Build builds and pushes the CoreDNS Docker image using buildx for multi-platform
// Note: buildx with push:true in config automatically pushes during build
func (Docker) Build() error {
	// Set git info as environment variables
	if err := setGitEnv(); err != nil {
		return err
	}

	if err := dockermagex.LoadConfig("docker.yaml"); err != nil {
		return err
	}
	// BuildxBuild with push:true in yaml will build AND push
	return dockermagex.BuildxBuild()
}

// Push pushes the CoreDNS Docker image to GHCR
func (Docker) Push() error {
	if err := dockermagex.LoadConfig("docker.yaml"); err != nil {
		return err
	}
	return dockermagex.Push()
}

// Login logs in to GitHub Container Registry
func (Docker) Login() error {
	if err := dockermagex.LoadConfig("docker.yaml"); err != nil {
		return err
	}
	return dockermagex.Login()
}

// Pull pulls the CoreDNS Docker image from GHCR
func (Docker) Pull() error {
	if err := dockermagex.LoadConfig("docker.yaml"); err != nil {
		return err
	}
	return dockermagex.Pull()
}

// BuildAndPush is an alias for Build (buildx already pushes with push:true)
func (d Docker) BuildAndPush() error {
	return d.Build()
}

// Info shows Git information
func (Git) Info() error {
	commitSHA, _ := gitmagex.GetCommitSHA()
	shortSHA, _ := gitmagex.GetShortCommitSHA()
	branch, _ := gitmagex.GetBranch()
	tag, _ := gitmagex.GetTag()
	version, _ := gitmagex.GetVersion()
	dirty, _ := gitmagex.IsDirty()

	fmt.Printf("Commit SHA:       %s\n", commitSHA)
	fmt.Printf("Short SHA:        %s\n", shortSHA)
	fmt.Printf("Branch:           %s\n", branch)
	fmt.Printf("Tag:              %s\n", tag)
	fmt.Printf("Version:          %s\n", version)
	fmt.Printf("Dirty:            %v\n", dirty)

	return nil
}

// setGitEnv sets git information as environment variables for Docker build
func setGitEnv() error {
	commitSHA, err := gitmagex.GetCommitSHA()
	if err != nil {
		return fmt.Errorf("failed to get commit SHA: %w", err)
	}

	version, err := gitmagex.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	os.Setenv("COMMIT_ID", commitSHA)
	os.Setenv("IMAGE_VERSION", version)

	fmt.Printf("🔧 Set COMMIT_ID=%s\n", commitSHA)
	fmt.Printf("🔧 Set IMAGE_VERSION=%s\n", version)

	return nil
}

// Made with Bob
