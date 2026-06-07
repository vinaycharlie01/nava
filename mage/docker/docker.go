package dockermagex

import (
	"github.com/nirantaraai/nava/pkg/docker"
)

// Package-level convenience functions for mage targets
var defaultRunner = dockerx.NewDockerRunner()

// LoadConfig loads Docker configuration from a YAML file
func LoadConfig(filepath string) error {
	return defaultRunner.LoadConfig(filepath)
}

// NewRunnerFromYAML creates a new runner with configuration loaded from YAML
func NewRunnerFromYAML(filepath string) (*dockerx.DockerRunner, error) {
	return dockerx.NewDockerRunnerFromYAML(filepath)
}

// Build builds a Docker image (requires loaded config)
func Build() error {
	return defaultRunner.Build()
}

// Push pushes a Docker image to a registry (requires loaded config)
func Push() error {
	return defaultRunner.Push()
}

// Pull pulls a Docker image from a registry (requires loaded config)
func Pull() error {
	return defaultRunner.Pull()
}

// Save saves one or more images to a tar archive (requires loaded config)
func Save() error {
	return defaultRunner.Save()
}

// Load loads an image from a tar archive (requires loaded config)
func Load() error {
	return defaultRunner.Load()
}

// Tag creates a tag for a source image (requires loaded config)
func Tag() error {
	return defaultRunner.Tag()
}

// RemoveImage removes one or more images (requires loaded config)
func RemoveImage() error {
	return defaultRunner.RemoveImage()
}

// ListImages lists Docker images (requires loaded config)
func ListImages() error {
	return defaultRunner.ListImages()
}

// Inspect returns low-level information on Docker objects (requires loaded config)
func Inspect() error {
	return defaultRunner.Inspect()
}

// Login logs in to a Docker registry (requires loaded config)
func Login() error {
	return defaultRunner.Login()
}

// Logout logs out from a Docker registry (requires loaded config)
func Logout() error {
	return defaultRunner.Logout()
}

// Prune removes unused Docker data (requires loaded config)
func Prune() error {
	return defaultRunner.Prune()
}

// BuildxBuild builds multi-platform images using buildx (requires loaded config)
func BuildxBuild() error {
	return defaultRunner.BuildxBuild()
}

// ComposeUp creates and starts containers (requires loaded config)
func ComposeUp() error {
	return defaultRunner.ComposeUp()
}

// ComposeDown stops and removes containers, networks (requires loaded config)
func ComposeDown() error {
	return defaultRunner.ComposeDown()
}

// ComposeBuild builds or rebuilds services (requires loaded config)
func ComposeBuild() error {
	return defaultRunner.ComposeBuild()
}

// Run runs a command in a new container (requires loaded config)
func Run() error {
	return defaultRunner.Run()
}

// Exec runs a command in a running container (requires loaded config)
func Exec() error {
	return defaultRunner.Exec()
}

// Logs fetches logs from a container (requires loaded config)
func Logs() error {
	return defaultRunner.Logs()
}

// Stop stops one or more running containers (requires loaded config)
func Stop() error {
	return defaultRunner.Stop()
}

// Start starts one or more stopped containers (requires loaded config)
func Start() error {
	return defaultRunner.Start()
}

// Restart restarts one or more containers (requires loaded config)
func Restart() error {
	return defaultRunner.Restart()
}

// Remove removes one or more containers (requires loaded config)
func Remove() error {
	return defaultRunner.Remove()
}

// Network manages Docker networks (requires loaded config)
func Network() error {
	return defaultRunner.Network()
}

// Volume manages Docker volumes (requires loaded config)
func Volume() error {
	return defaultRunner.Volume()
}

// Made with Bob
