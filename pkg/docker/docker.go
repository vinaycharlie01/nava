package dockerx

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

// DockerExecutor defines the interface for Docker operations
type DockerExecutor interface {
	LoadConfig(filepath string) error
	Build() error
	Push() error
	Pull() error
	Save() error
	Load() error
	Tag() error
	RemoveImage() error
	ListImages() error
	Inspect() error
	Login() error
	Logout() error
	Prune() error
	BuildxBuild() error
	ComposeUp() error
	ComposeDown() error
	ComposeBuild() error
	Run() error
	Exec() error
	Logs() error
	Stop() error
	Start() error
	Restart() error
	Remove() error
	Network() error
	Volume() error
}

// DockerRunner handles docker command execution with dependency injection
type DockerRunner struct {
	executor execx.Executor
	config   *DockerConfig
}

// NewDockerRunner creates a new DockerRunner with the default executor
func NewDockerRunner() *DockerRunner {
	return &DockerRunner{
		executor: execx.NewExec(),
	}
}

// NewDockerRunnerWithExecutor creates a new DockerRunner with a custom executor
func NewDockerRunnerWithExecutor(executor execx.Executor) *DockerRunner {
	return &DockerRunner{
		executor: executor,
	}
}

// NewDockerRunnerFromYAML creates a new DockerRunner with configuration loaded from YAML
func NewDockerRunnerFromYAML(filepath string) (*DockerRunner, error) {
	config, err := LoadDockerConfig(filepath)
	if err != nil {
		return nil, err
	}
	return &DockerRunner{
		executor: execx.NewExec(),
		config:   config,
	}, nil
}

// LoadConfig loads Docker configuration from a YAML file
func (d *DockerRunner) LoadConfig(filepath string) error {
	config, err := LoadDockerConfig(filepath)
	if err != nil {
		return err
	}
	d.config = config
	return nil
}

// DockerConfig contains all Docker operation configurations
type DockerConfig struct {
	Build       *BuildOptions       `yaml:"build,omitempty"`
	BuildxBuild *BuildxBuildOptions `yaml:"buildxBuild,omitempty"`
	Push        *PushOptions        `yaml:"push,omitempty"`
	Pull        *PullOptions        `yaml:"pull,omitempty"`
	Save        *SaveOptions        `yaml:"save,omitempty"`
	Load        *LoadOptions        `yaml:"load,omitempty"`
	Tag         *TagOptions         `yaml:"tag,omitempty"`
	RemoveImage *RemoveImageOptions `yaml:"removeImage,omitempty"`
	ListImages  *ListImagesOptions  `yaml:"listImages,omitempty"`
	Inspect     *InspectOptions     `yaml:"inspect,omitempty"`
	Login       *LoginOptions       `yaml:"login,omitempty"`
	Logout      *LogoutOptions      `yaml:"logout,omitempty"`
	Prune       *PruneOptions       `yaml:"prune,omitempty"`
	Compose     *ComposeOptions     `yaml:"compose,omitempty"`
	Run         *RunOptions         `yaml:"run,omitempty"`
	Exec        *ExecOptions        `yaml:"exec,omitempty"`
	Logs        *LogsOptions        `yaml:"logs,omitempty"`
	Stop        *StopOptions        `yaml:"stop,omitempty"`
	Start       *StartOptions       `yaml:"start,omitempty"`
	Restart     *RestartOptions     `yaml:"restart,omitempty"`
	Remove      *RemoveOptions      `yaml:"remove,omitempty"`
	Network     *NetworkOptions     `yaml:"network,omitempty"`
	Volume      *VolumeOptions      `yaml:"volume,omitempty"`
}

// LoadDockerConfig loads Docker configuration from a YAML file
func LoadDockerConfig(filepath string) (*DockerConfig, error) {
	var config DockerConfig

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// BuildOptions contains options for docker build
type BuildOptions struct {
	Context    string            `yaml:"context"`    // Build context path (default: ".")
	Dockerfile string            `yaml:"dockerfile"` // Path to Dockerfile (default: "Dockerfile")
	Tags       []string          `yaml:"tags"`       // Image tags
	BuildArgs  map[string]string `yaml:"buildArgs"`  // Build arguments
	Target     string            `yaml:"target"`     // Target build stage
	Platform   []string          `yaml:"platform"`   // Target platforms (e.g., linux/amd64,linux/arm64)
	NoCache    bool              `yaml:"noCache"`    // Do not use cache when building
	Pull       bool              `yaml:"pull"`       // Always attempt to pull newer version of base image
	Quiet      bool              `yaml:"quiet"`      // Suppress build output
	Labels     map[string]string `yaml:"labels"`     // Set metadata for an image
	CacheFrom  []string          `yaml:"cacheFrom"`  // Images to consider as cache sources
	Network    string            `yaml:"network"`    // Set the networking mode for RUN instructions
	Progress   string            `yaml:"progress"`   // Set type of progress output (auto, plain, tty)
}

// Build builds a Docker image using loaded configuration
func (d *DockerRunner) Build() error {
	if d.config == nil || d.config.Build == nil {
		return fmt.Errorf("build configuration not loaded")
	}

	opts := d.config.Build
	if len(opts.Tags) == 0 {
		return fmt.Errorf("at least one tag is required")
	}

	buildContext := opts.Context
	if buildContext == "" {
		buildContext = "."
	}

	slog.Info("🐳 Building Docker image...",
		"tags", opts.Tags,
		"context", buildContext,
		"dockerfile", opts.Dockerfile,
	)

	start := time.Now()
	args := []string{"build"}

	for _, tag := range opts.Tags {
		args = append(args, "-t", tag)
	}

	if opts.Dockerfile != "" {
		args = append(args, "-f", opts.Dockerfile)
	}

	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	for _, platform := range opts.Platform {
		args = append(args, "--platform", platform)
	}

	for key, value := range opts.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	for _, cache := range opts.CacheFrom {
		args = append(args, "--cache-from", cache)
	}

	if opts.Network != "" {
		args = append(args, "--network", opts.Network)
	}

	if opts.Progress != "" {
		args = append(args, "--progress", opts.Progress)
	}

	if opts.NoCache {
		args = append(args, "--no-cache")
	}

	if opts.Pull {
		args = append(args, "--pull")
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	args = append(args, buildContext)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker image built successfully", "duration", time.Since(start))
	return nil
}

// PushOptions contains options for docker push
type PushOptions struct {
	Image         string `yaml:"image"`         // Image name with tag
	AllTags       bool   `yaml:"allTags"`       // Push all tags
	Quiet         bool   `yaml:"quiet"`         // Suppress verbose output
	DisableDigest bool   `yaml:"disableDigest"` // Don't print image digest after push
}

// Push pushes a Docker image to a registry
func (d *DockerRunner) Push() error {
	if d.config == nil || d.config.Push == nil {
		return fmt.Errorf("push configuration not loaded")
	}

	opts := d.config.Push
	if opts.Image == "" {
		return fmt.Errorf("image name is required")
	}

	slog.Info("📤 Pushing Docker image...", "image", opts.Image)
	start := time.Now()

	args := []string{"push"}

	if opts.AllTags {
		args = append(args, "--all-tags")
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	if opts.DisableDigest {
		args = append(args, "--disable-content-trust")
	}

	args = append(args, opts.Image)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker image pushed successfully", "duration", time.Since(start))
	return nil
}

// PullOptions contains options for docker pull
type PullOptions struct {
	Image    string `yaml:"image"`    // Image name with tag
	Platform string `yaml:"platform"` // Set platform if server is multi-platform capable
	AllTags  bool   `yaml:"allTags"`  // Download all tagged images in the repository
	Quiet    bool   `yaml:"quiet"`    // Suppress verbose output
}

// Pull pulls a Docker image from a registry
func (d *DockerRunner) Pull() error {
	if d.config == nil || d.config.Pull == nil {
		return fmt.Errorf("pull configuration not loaded")
	}

	opts := d.config.Pull
	if opts.Image == "" {
		return fmt.Errorf("image name is required")
	}

	slog.Info("⬇️  Pulling Docker image...", "image", opts.Image)
	start := time.Now()

	args := []string{"pull"}

	if opts.Platform != "" {
		args = append(args, "--platform", opts.Platform)
	}

	if opts.AllTags {
		args = append(args, "--all-tags")
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	args = append(args, opts.Image)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker image pulled successfully", "duration", time.Since(start))
	return nil
}

// SaveOptions contains options for docker save
type SaveOptions struct {
	Images []string `yaml:"images"` // Image names to save
	Output string   `yaml:"output"` // Write to a file instead of STDOUT
}

// Save saves one or more images to a tar archive
func (d *DockerRunner) Save() error {
	if d.config == nil || d.config.Save == nil {
		return fmt.Errorf("save configuration not loaded")
	}

	opts := d.config.Save
	if len(opts.Images) == 0 {
		return fmt.Errorf("at least one image is required")
	}

	if opts.Output == "" {
		return fmt.Errorf("output file is required")
	}

	slog.Info("💾 Saving Docker images...", "images", opts.Images, "output", opts.Output)
	start := time.Now()

	args := []string{"save", "-o", opts.Output}
	args = append(args, opts.Images...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker images saved successfully", "duration", time.Since(start))
	return nil
}

// LoadOptions contains options for docker load
type LoadOptions struct {
	Input string `yaml:"input"` // Read from tar archive file instead of STDIN
	Quiet bool   `yaml:"quiet"` // Suppress the load output
}

// Load loads an image from a tar archive
func (d *DockerRunner) Load() error {
	if d.config == nil || d.config.Load == nil {
		return fmt.Errorf("load configuration not loaded")
	}

	opts := d.config.Load
	if opts.Input == "" {
		return fmt.Errorf("input file is required")
	}

	slog.Info("📥 Loading Docker image...", "input", opts.Input)
	start := time.Now()

	args := []string{"load", "-i", opts.Input}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker image loaded successfully", "duration", time.Since(start))
	return nil
}

// TagOptions contains options for docker tag
type TagOptions struct {
	SourceImage string `yaml:"sourceImage"` // Source image name
	TargetImage string `yaml:"targetImage"` // Target image name with tag
}

// Tag creates a tag for a source image
func (d *DockerRunner) Tag() error {
	if d.config == nil || d.config.Tag == nil {
		return fmt.Errorf("tag configuration not loaded")
	}

	opts := d.config.Tag
	if opts.SourceImage == "" {
		return fmt.Errorf("source image is required")
	}

	if opts.TargetImage == "" {
		return fmt.Errorf("target image is required")
	}

	slog.Info("🏷️  Tagging Docker image...",
		"source", opts.SourceImage,
		"target", opts.TargetImage,
	)
	start := time.Now()

	args := []string{"tag", opts.SourceImage, opts.TargetImage}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker image tagged successfully", "duration", time.Since(start))
	return nil
}

// RemoveImageOptions contains options for docker rmi
type RemoveImageOptions struct {
	Images  []string `yaml:"images"`  // Image names to remove
	Force   bool     `yaml:"force"`   // Force removal of the image
	NoPrune bool     `yaml:"noPrune"` // Do not delete untagged parents
}

// RemoveImage removes one or more images
func (d *DockerRunner) RemoveImage() error {
	if d.config == nil || d.config.RemoveImage == nil {
		return fmt.Errorf("removeImage configuration not loaded")
	}

	opts := d.config.RemoveImage
	if len(opts.Images) == 0 {
		return fmt.Errorf("at least one image is required")
	}

	slog.Info("🗑️  Removing Docker images...", "images", opts.Images)
	start := time.Now()

	args := []string{"rmi"}

	if opts.Force {
		args = append(args, "--force")
	}

	if opts.NoPrune {
		args = append(args, "--no-prune")
	}

	args = append(args, opts.Images...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker images removed successfully", "duration", time.Since(start))
	return nil
}

// ListImagesOptions contains options for docker images
type ListImagesOptions struct {
	All     bool     `yaml:"all"`     // Show all images (default hides intermediate images)
	Digests bool     `yaml:"digests"` // Show digests
	Filter  []string `yaml:"filter"`  // Filter output based on conditions
	Format  string   `yaml:"format"`  // Pretty-print images using a Go template
	NoTrunc bool     `yaml:"noTrunc"` // Don't truncate output
	Quiet   bool     `yaml:"quiet"`   // Only show image IDs
}

// ListImages lists Docker images
func (d *DockerRunner) ListImages() error {
	if d.config == nil || d.config.ListImages == nil {
		return fmt.Errorf("listImages configuration not loaded")
	}

	opts := d.config.ListImages
	slog.Info("📋 Listing Docker images...")
	start := time.Now()

	args := []string{"images"}

	if opts.All {
		args = append(args, "--all")
	}

	if opts.Digests {
		args = append(args, "--digests")
	}

	for _, filter := range opts.Filter {
		args = append(args, "--filter", filter)
	}

	if opts.Format != "" {
		args = append(args, "--format", opts.Format)
	}

	if opts.NoTrunc {
		args = append(args, "--no-trunc")
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker images listed", "duration", time.Since(start))
	return nil
}

// InspectOptions contains options for docker inspect
type InspectOptions struct {
	Targets []string `yaml:"targets"` // Image or container names/IDs to inspect
	Format  string   `yaml:"format"`  // Format the output using a Go template
	Type    string   `yaml:"type"`    // Return JSON for specified type (image, container, etc.)
}

// Inspect returns low-level information on Docker objects
func (d *DockerRunner) Inspect() error {
	if d.config == nil || d.config.Inspect == nil {
		return fmt.Errorf("inspect configuration not loaded")
	}

	opts := d.config.Inspect
	if len(opts.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	slog.Info("🔍 Inspecting Docker objects...", "targets", opts.Targets)
	start := time.Now()

	args := []string{"inspect"}

	if opts.Format != "" {
		args = append(args, "--format", opts.Format)
	}

	if opts.Type != "" {
		args = append(args, "--type", opts.Type)
	}

	args = append(args, opts.Targets...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker objects inspected", "duration", time.Since(start))
	return nil
}

// LoginOptions contains options for docker login
type LoginOptions struct {
	Registry      string `yaml:"registry"`      // Registry server (default: Docker Hub)
	Username      string `yaml:"username"`      // Username
	Password      string `yaml:"password"`      // Password
	PasswordStdin bool   `yaml:"passwordStdin"` // Take password from stdin
}

// Login logs in to a Docker registry
func (d *DockerRunner) Login() error {
	if d.config == nil || d.config.Login == nil {
		return fmt.Errorf("login configuration not loaded")
	}

	opts := d.config.Login
	slog.Info("🔐 Logging in to Docker registry...", "registry", opts.Registry)
	start := time.Now()

	args := []string{"login"}

	if opts.Username != "" {
		args = append(args, "--username", opts.Username)
	}

	if opts.Password != "" {
		args = append(args, "--password", opts.Password)
	}

	if opts.PasswordStdin {
		args = append(args, "--password-stdin")
	}

	if opts.Registry != "" {
		args = append(args, opts.Registry)
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Logged in successfully", "duration", time.Since(start))
	return nil
}

// LogoutOptions contains options for docker logout
type LogoutOptions struct {
	Registry string `yaml:"registry"` // Registry server (default: Docker Hub)
}

// Logout logs out from a Docker registry
func (d *DockerRunner) Logout() error {
	if d.config == nil || d.config.Logout == nil {
		return fmt.Errorf("logout configuration not loaded")
	}

	opts := d.config.Logout
	slog.Info("🔓 Logging out from Docker registry...", "registry", opts.Registry)
	start := time.Now()

	args := []string{"logout"}

	if opts.Registry != "" {
		args = append(args, opts.Registry)
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Logged out successfully", "duration", time.Since(start))
	return nil
}

// PruneOptions contains options for docker system prune
type PruneOptions struct {
	All     bool     `yaml:"all"`     // Remove all unused images not just dangling ones
	Volumes bool     `yaml:"volumes"` // Prune volumes
	Force   bool     `yaml:"force"`   // Do not prompt for confirmation
	Filter  []string `yaml:"filter"`  // Provide filter values
}

// Prune removes unused Docker data
func (d *DockerRunner) Prune() error {
	if d.config == nil || d.config.Prune == nil {
		return fmt.Errorf("prune configuration not loaded")
	}

	opts := d.config.Prune
	slog.Info("🧹 Pruning Docker system...")
	start := time.Now()

	args := []string{"system", "prune"}

	if opts.All {
		args = append(args, "--all")
	}

	if opts.Volumes {
		args = append(args, "--volumes")
	}

	if opts.Force {
		args = append(args, "--force")
	}

	for _, filter := range opts.Filter {
		args = append(args, "--filter", filter)
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker system pruned", "duration", time.Since(start))
	return nil
}

// BuildxBuildOptions contains options for docker buildx build (multi-platform builds)
type BuildxBuildOptions struct {
	Context    string            `yaml:"context"`    // Build context path
	Dockerfile string            `yaml:"dockerfile"` // Path to Dockerfile
	Tags       []string          `yaml:"tags"`       // Image tags
	BuildArgs  map[string]string `yaml:"buildArgs"`  // Build arguments
	Target     string            `yaml:"target"`     // Target build stage
	Platforms  []string          `yaml:"platforms"`  // Target platforms (e.g., linux/amd64,linux/arm64)
	Push       bool              `yaml:"push"`       // Push to registry after build
	Load       bool              `yaml:"load"`       // Load the single-platform build result to docker images
	NoCache    bool              `yaml:"noCache"`    // Do not use cache when building
	Pull       bool              `yaml:"pull"`       // Always attempt to pull newer version of base image
	Builder    string            `yaml:"builder"`    // Override the configured builder instance
	CacheFrom  []string          `yaml:"cacheFrom"`  // External cache sources
	CacheTo    []string          `yaml:"cacheTo"`    // Cache export destinations
	Output     string            `yaml:"output"`     // Output destination (e.g., type=docker,dest=./image.tar)
}

// BuildxBuild builds multi-platform images using buildx
func (d *DockerRunner) BuildxBuild() error {
	if d.config == nil || d.config.BuildxBuild == nil {
		return fmt.Errorf("buildxBuild configuration not loaded")
	}

	opts := d.config.BuildxBuild
	if len(opts.Tags) == 0 {
		return fmt.Errorf("at least one tag is required")
	}

	buildContext := opts.Context
	if buildContext == "" {
		buildContext = "."
	}

	slog.Info("🏗️  Building multi-platform Docker image with buildx...",
		"tags", opts.Tags,
		"platforms", opts.Platforms,
	)
	start := time.Now()

	args := []string{"buildx", "build"}

	for _, tag := range opts.Tags {
		args = append(args, "-t", tag)
	}

	if opts.Dockerfile != "" {
		args = append(args, "-f", opts.Dockerfile)
	}

	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	if len(opts.Platforms) > 0 {
		args = append(args, "--platform", strings.Join(opts.Platforms, ","))
	}

	if opts.Builder != "" {
		args = append(args, "--builder", opts.Builder)
	}

	for _, cache := range opts.CacheFrom {
		args = append(args, "--cache-from", cache)
	}

	for _, cache := range opts.CacheTo {
		args = append(args, "--cache-to", cache)
	}

	if opts.Output != "" {
		args = append(args, "--output", opts.Output)
	}

	if opts.Push {
		args = append(args, "--push")
	}

	if opts.Load {
		args = append(args, "--load")
	}

	if opts.NoCache {
		args = append(args, "--no-cache")
	}

	if opts.Pull {
		args = append(args, "--pull")
	}

	args = append(args, buildContext)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Multi-platform Docker image built successfully", "duration", time.Since(start))
	return nil
}

// ComposeOptions contains options for docker compose commands
type ComposeOptions struct {
	Files       []string `yaml:"files"`       // Compose file paths
	ProjectName string   `yaml:"projectName"` // Project name
	EnvFile     string   `yaml:"envFile"`     // Environment file
	Profiles    []string `yaml:"profiles"`    // Profiles to enable
	Services    []string `yaml:"services"`    // Services to operate on
	Build       bool     `yaml:"build"`       // Build images before starting
	Detach      bool     `yaml:"detach"`      // Detached mode
	Remove      bool     `yaml:"remove"`      // Remove containers for services not defined
	Volumes     bool     `yaml:"volumes"`     // Remove named volumes
	Force       bool     `yaml:"force"`       // Force operation
}

// ComposeUp creates and starts containers
func (d *DockerRunner) ComposeUp() error {
	if d.config == nil || d.config.Compose == nil {
		return fmt.Errorf("compose configuration not loaded")
	}

	opts := d.config.Compose
	slog.Info("🚀 Starting Docker Compose services...", "services", opts.Services)
	start := time.Now()

	args := d.buildComposeArgs(opts, "up")

	if opts.Build {
		args = append(args, "--build")
	}

	if opts.Detach {
		args = append(args, "--detach")
	}

	if opts.Remove {
		args = append(args, "--remove-orphans")
	}

	args = append(args, opts.Services...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker Compose services started", "duration", time.Since(start))
	return nil
}

// ComposeDown stops and removes containers, networks
func (d *DockerRunner) ComposeDown() error {
	if d.config == nil || d.config.Compose == nil {
		return fmt.Errorf("compose configuration not loaded")
	}

	opts := d.config.Compose
	slog.Info("🛑 Stopping Docker Compose services...")
	start := time.Now()

	args := d.buildComposeArgs(opts, "down")

	if opts.Volumes {
		args = append(args, "--volumes")
	}

	if opts.Remove {
		args = append(args, "--remove-orphans")
	}

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker Compose services stopped", "duration", time.Since(start))
	return nil
}

// ComposeBuild builds or rebuilds services
func (d *DockerRunner) ComposeBuild() error {
	if d.config == nil || d.config.Compose == nil {
		return fmt.Errorf("compose configuration not loaded")
	}

	opts := d.config.Compose
	slog.Info("🔨 Building Docker Compose services...", "services", opts.Services)
	start := time.Now()

	args := d.buildComposeArgs(opts, "build")
	args = append(args, opts.Services...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker Compose services built", "duration", time.Since(start))
	return nil
}

// buildComposeArgs builds common compose arguments
func (d *DockerRunner) buildComposeArgs(opts *ComposeOptions, command string) []string {
	args := []string{"compose"}

	for _, file := range opts.Files {
		args = append(args, "-f", file)
	}

	if opts.ProjectName != "" {
		args = append(args, "-p", opts.ProjectName)
	}

	if opts.EnvFile != "" {
		args = append(args, "--env-file", opts.EnvFile)
	}

	for _, profile := range opts.Profiles {
		args = append(args, "--profile", profile)
	}

	args = append(args, command)
	return args
}

// RunOptions contains options for docker run
type RunOptions struct {
	Image       string            `yaml:"image"`       // Image to run
	Name        string            `yaml:"name"`        // Container name
	Detach      bool              `yaml:"detach"`      // Run container in background
	Remove      bool              `yaml:"remove"`      // Automatically remove container when it exits
	Interactive bool              `yaml:"interactive"` // Keep STDIN open
	TTY         bool              `yaml:"tty"`         // Allocate a pseudo-TTY
	Env         map[string]string `yaml:"env"`         // Environment variables
	Ports       []string          `yaml:"ports"`       // Port mappings (host:container)
	Volumes     []string          `yaml:"volumes"`     // Volume mounts (host:container)
	Network     string            `yaml:"network"`     // Network to connect to
	Command     []string          `yaml:"command"`     // Command to run
}

// Run runs a command in a new container
func (d *DockerRunner) Run() error {
	if d.config == nil || d.config.Run == nil {
		return fmt.Errorf("run configuration not loaded")
	}

	opts := d.config.Run
	if opts.Image == "" {
		return fmt.Errorf("image is required")
	}

	slog.Info("🏃 Running Docker container...", "image", opts.Image, "name", opts.Name)
	start := time.Now()

	args := []string{"run"}

	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}

	if opts.Detach {
		args = append(args, "-d")
	}

	if opts.Remove {
		args = append(args, "--rm")
	}

	if opts.Interactive {
		args = append(args, "-i")
	}

	if opts.TTY {
		args = append(args, "-t")
	}

	for key, value := range opts.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	for _, port := range opts.Ports {
		args = append(args, "-p", port)
	}

	for _, volume := range opts.Volumes {
		args = append(args, "-v", volume)
	}

	if opts.Network != "" {
		args = append(args, "--network", opts.Network)
	}

	args = append(args, opts.Image)
	args = append(args, opts.Command...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Docker container started", "duration", time.Since(start))
	return nil
}

// ExecOptions contains options for docker exec
type ExecOptions struct {
	Container   string   `yaml:"container"`   // Container name or ID
	Command     []string `yaml:"command"`     // Command to execute
	Detach      bool     `yaml:"detach"`      // Detached mode
	Interactive bool     `yaml:"interactive"` // Keep STDIN open
	TTY         bool     `yaml:"tty"`         // Allocate a pseudo-TTY
	User        string   `yaml:"user"`        // Username or UID
	WorkDir     string   `yaml:"workDir"`     // Working directory
}

// Exec runs a command in a running container
func (d *DockerRunner) Exec() error {
	if d.config == nil || d.config.Exec == nil {
		return fmt.Errorf("exec configuration not loaded")
	}

	opts := d.config.Exec
	if opts.Container == "" {
		return fmt.Errorf("container is required")
	}

	if len(opts.Command) == 0 {
		return fmt.Errorf("command is required")
	}

	slog.Info("⚡ Executing command in container...", "container", opts.Container)
	start := time.Now()

	args := []string{"exec"}

	if opts.Detach {
		args = append(args, "-d")
	}

	if opts.Interactive {
		args = append(args, "-i")
	}

	if opts.TTY {
		args = append(args, "-t")
	}

	if opts.User != "" {
		args = append(args, "-u", opts.User)
	}

	if opts.WorkDir != "" {
		args = append(args, "-w", opts.WorkDir)
	}

	args = append(args, opts.Container)
	args = append(args, opts.Command...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Command executed", "duration", time.Since(start))
	return nil
}

// LogsOptions contains options for docker logs
type LogsOptions struct {
	Container  string `yaml:"container"`  // Container name or ID
	Follow     bool   `yaml:"follow"`     // Follow log output
	Timestamps bool   `yaml:"timestamps"` // Show timestamps
	Tail       string `yaml:"tail"`       // Number of lines to show from the end
	Since      string `yaml:"since"`      // Show logs since timestamp
	Until      string `yaml:"until"`      // Show logs before timestamp
}

// Logs fetches logs from a container
func (d *DockerRunner) Logs() error {
	if d.config == nil || d.config.Logs == nil {
		return fmt.Errorf("logs configuration not loaded")
	}

	opts := d.config.Logs
	if opts.Container == "" {
		return fmt.Errorf("container is required")
	}

	slog.Info("📜 Fetching container logs...", "container", opts.Container)
	start := time.Now()

	args := []string{"logs"}

	if opts.Follow {
		args = append(args, "-f")
	}

	if opts.Timestamps {
		args = append(args, "-t")
	}

	if opts.Tail != "" {
		args = append(args, "--tail", opts.Tail)
	}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	if opts.Until != "" {
		args = append(args, "--until", opts.Until)
	}

	args = append(args, opts.Container)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Logs fetched", "duration", time.Since(start))
	return nil
}

// StopOptions contains options for docker stop
type StopOptions struct {
	Containers []string `yaml:"containers"` // Container names or IDs
	Time       int      `yaml:"time"`       // Seconds to wait before killing
}

// Stop stops one or more running containers
func (d *DockerRunner) Stop() error {
	if d.config == nil || d.config.Stop == nil {
		return fmt.Errorf("stop configuration not loaded")
	}

	opts := d.config.Stop
	if len(opts.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	slog.Info("🛑 Stopping containers...", "containers", opts.Containers)
	start := time.Now()

	args := []string{"stop"}

	if opts.Time > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", opts.Time))
	}

	args = append(args, opts.Containers...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Containers stopped", "duration", time.Since(start))
	return nil
}

// StartOptions contains options for docker start
type StartOptions struct {
	Containers  []string `yaml:"containers"`  // Container names or IDs
	Attach      bool     `yaml:"attach"`      // Attach STDOUT/STDERR
	Interactive bool     `yaml:"interactive"` // Attach container's STDIN
}

// Start starts one or more stopped containers
func (d *DockerRunner) Start() error {
	if d.config == nil || d.config.Start == nil {
		return fmt.Errorf("start configuration not loaded")
	}

	opts := d.config.Start
	if len(opts.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	slog.Info("▶️  Starting containers...", "containers", opts.Containers)
	start := time.Now()

	args := []string{"start"}

	if opts.Attach {
		args = append(args, "-a")
	}

	if opts.Interactive {
		args = append(args, "-i")
	}

	args = append(args, opts.Containers...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Containers started", "duration", time.Since(start))
	return nil
}

// RestartOptions contains options for docker restart
type RestartOptions struct {
	Containers []string `yaml:"containers"` // Container names or IDs
	Time       int      `yaml:"time"`       // Seconds to wait before killing
}

// Restart restarts one or more containers
func (d *DockerRunner) Restart() error {
	if d.config == nil || d.config.Restart == nil {
		return fmt.Errorf("restart configuration not loaded")
	}

	opts := d.config.Restart
	if len(opts.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	slog.Info("🔄 Restarting containers...", "containers", opts.Containers)
	start := time.Now()

	args := []string{"restart"}

	if opts.Time > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", opts.Time))
	}

	args = append(args, opts.Containers...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Containers restarted", "duration", time.Since(start))
	return nil
}

// RemoveOptions contains options for docker rm
type RemoveOptions struct {
	Containers []string `yaml:"containers"` // Container names or IDs
	Force      bool     `yaml:"force"`      // Force removal
	Volumes    bool     `yaml:"volumes"`    // Remove associated volumes
}

// Remove removes one or more containers
func (d *DockerRunner) Remove() error {
	if d.config == nil || d.config.Remove == nil {
		return fmt.Errorf("remove configuration not loaded")
	}

	opts := d.config.Remove
	if len(opts.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	slog.Info("🗑️  Removing containers...", "containers", opts.Containers)
	start := time.Now()

	args := []string{"rm"}

	if opts.Force {
		args = append(args, "-f")
	}

	if opts.Volumes {
		args = append(args, "-v")
	}

	args = append(args, opts.Containers...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Containers removed", "duration", time.Since(start))
	return nil
}

// NetworkOptions contains options for docker network commands
type NetworkOptions struct {
	Command string   `yaml:"command"` // Network command (create, rm, ls, inspect, connect, disconnect)
	Name    string   `yaml:"name"`    // Network name
	Driver  string   `yaml:"driver"`  // Network driver
	Args    []string `yaml:"args"`    // Additional arguments
}

// Network manages Docker networks
func (d *DockerRunner) Network() error {
	if d.config == nil || d.config.Network == nil {
		return fmt.Errorf("network configuration not loaded")
	}

	opts := d.config.Network
	if opts.Command == "" {
		return fmt.Errorf("network command is required")
	}

	slog.Info("🌐 Managing Docker network...", "command", opts.Command, "name", opts.Name)
	start := time.Now()

	args := []string{"network", opts.Command}

	if opts.Driver != "" {
		args = append(args, "--driver", opts.Driver)
	}

	if opts.Name != "" {
		args = append(args, opts.Name)
	}

	args = append(args, opts.Args...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Network operation completed", "duration", time.Since(start))
	return nil
}

// VolumeOptions contains options for docker volume commands
type VolumeOptions struct {
	Command string   `yaml:"command"` // Volume command (create, rm, ls, inspect, prune)
	Name    string   `yaml:"name"`    // Volume name
	Driver  string   `yaml:"driver"`  // Volume driver
	Args    []string `yaml:"args"`    // Additional arguments
}

// Volume manages Docker volumes
func (d *DockerRunner) Volume() error {
	if d.config == nil || d.config.Volume == nil {
		return fmt.Errorf("volume configuration not loaded")
	}

	opts := d.config.Volume
	if opts.Command == "" {
		return fmt.Errorf("volume command is required")
	}

	slog.Info("💾 Managing Docker volume...", "command", opts.Command, "name", opts.Name)
	start := time.Now()

	args := []string{"volume", opts.Command}

	if opts.Driver != "" {
		args = append(args, "--driver", opts.Driver)
	}

	if opts.Name != "" {
		args = append(args, opts.Name)
	}

	args = append(args, opts.Args...)

	if err := d.executor.Run(context.Background(), "docker", false, args...); err != nil {
		return err
	}

	slog.Info("✅ Volume operation completed", "duration", time.Since(start))
	return nil
}

// Made with Bob
