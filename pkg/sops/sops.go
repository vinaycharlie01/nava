package sopsx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/nirantaraai/nava/pkg/exec"

	"gopkg.in/yaml.v3"
)

// SopsConfig is the root YAML configuration for SOPS secrets.
type SopsConfig struct {
	Secrets []string `yaml:"secrets"`
}

// SopsRunner handles SOPS encrypt/decrypt operations.
type SopsRunner struct {
	executor execx.Executor
	config   *SopsConfig
}

// NewSopsRunner creates a new SopsRunner with default executor.
func NewSopsRunner() *SopsRunner {
	return &SopsRunner{
		executor: execx.NewExec(),
	}
}

// NewSopsRunnerWithExecutor creates a SopsRunner with a custom executor.
func NewSopsRunnerWithExecutor(executor execx.Executor) *SopsRunner {
	return &SopsRunner{
		executor: executor,
	}
}

// NewSopsRunnerFromYAML creates a runner with configuration loaded from YAML.
func NewSopsRunnerFromYAML(filepath string) (*SopsRunner, error) {
	runner := NewSopsRunner()
	if err := runner.LoadConfig(filepath); err != nil {
		return nil, err
	}
	return runner, nil
}

// LoadConfig loads SOPS secret paths from a YAML config file.
// It reads the top-level "secrets" key from the given file (e.g. k3d.yaml).
func (s *SopsRunner) LoadConfig(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read sops config: %w", err)
	}
	var cfg SopsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse sops config: %w", err)
	}
	s.config = &cfg
	return nil
}

// Config returns the loaded configuration.
func (s *SopsRunner) Config() *SopsConfig {
	return s.config
}

// SecretPaths returns the configured secret paths.
func (s *SopsRunner) SecretPaths() []string {
	if s.config == nil {
		return nil
	}
	return s.config.Secrets
}

// EnsureAgeKeyFile sets SOPS_AGE_KEY_FILE if not already set, auto-detecting ~/.sops/age-key.txt.
func (s *SopsRunner) EnsureAgeKeyFile() error {
	if os.Getenv("SOPS_AGE_KEY_FILE") != "" {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home dir: %w", err)
	}
	keyFile := home + "/.sops/age-key.txt"
	if _, err := os.Stat(keyFile); err != nil {
		return fmt.Errorf("SOPS_AGE_KEY_FILE not set and no key found at %s. Run 'mage sopsInit' first", keyFile)
	}
	slog.Info("🔑 Auto-detected age key", "path", keyFile)
	os.Setenv("SOPS_AGE_KEY_FILE", keyFile)
	return nil
}

// DecryptSecrets decrypts all .enc.yaml files to .dec.yaml on disk.
func (s *SopsRunner) DecryptSecrets() error {
	if len(s.config.Secrets) == 0 {
		slog.Info("⏭️  No secrets configured, skipping")
		return nil
	}

	if err := s.EnsureAgeKeyFile(); err != nil {
		return err
	}

	for _, encPath := range s.config.Secrets {
		decPath := strings.TrimSuffix(encPath, ".enc.yaml") + ".dec.yaml"
		slog.Info("🔐 Decrypting secret", "src", encPath, "dst", decPath)

		cmd := exec.CommandContext(context.Background(), "sops", "-d", encPath)
		cmd.Env = append(os.Environ(), "SOPS_AGE_KEY_FILE="+os.Getenv("SOPS_AGE_KEY_FILE"))
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("sops decrypt %s failed: %w", encPath, err)
		}
		if err := os.WriteFile(decPath, out, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", decPath, err)
		}
	}

	slog.Info("✅ Secrets decrypted")
	return nil
}

// EncryptSecrets encrypts all .dec.yaml files to .enc.yaml.
func (s *SopsRunner) EncryptSecrets() error {
	if len(s.config.Secrets) == 0 {
		slog.Info("⏭️  No secrets configured, skipping")
		return nil
	}

	if err := s.EnsureAgeKeyFile(); err != nil {
		return err
	}

	for _, encPath := range s.config.Secrets {
		decPath := strings.TrimSuffix(encPath, ".enc.yaml") + ".dec.yaml"
		slog.Info("🔒 Encrypting secret", "src", decPath, "dst", encPath)

		out, err := exec.CommandContext(context.Background(), "sops", "-e", decPath).Output()
		if err != nil {
			return fmt.Errorf("sops encrypt %s failed: %w", decPath, err)
		}
		if err := os.WriteFile(encPath, out, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", encPath, err)
		}
	}

	slog.Info("✅ Secrets encrypted")
	return nil
}

// InstallTools installs sops and age via go install.
func (s *SopsRunner) InstallTools() error {
	tools := []struct {
		name string
		pkg  string
	}{
		{"age", "filippo.io/age/cmd/...@latest"},
		{"sops", "github.com/getsops/sops/v3/cmd/sops@latest"},
	}

	for _, t := range tools {
		slog.Info("📦 Installing", "tool", t.name)
		if err := s.executor.Run(context.Background(), "go", false, "install", t.pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", t.name, err)
		}
	}

	slog.Info("✅ sops and age installed")
	return nil
}

// Init installs tools, generates an age key (if missing), updates .sops.yaml, and encrypts all secrets.
func (s *SopsRunner) Init() error {
	slog.Info("--- InstallTools ---")
	if err := s.InstallTools(); err != nil {
		return fmt.Errorf("InstallTools failed: %w", err)
	}

	slog.Info("--- GenerateAgeKey ---")
	if err := s.GenerateAgeKey(); err != nil {
		return fmt.Errorf("GenerateAgeKey failed: %w", err)
	}

	slog.Info("--- EncryptSecrets ---")
	if err := s.EncryptSecrets(); err != nil {
		return fmt.Errorf("EncryptSecrets failed: %w", err)
	}

	return nil
}

// GenerateAgeKey generates an age key at ~/.sops/age-key.txt if it doesn't exist,
// and updates .sops.yaml with the public key.
func (s *SopsRunner) GenerateAgeKey() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home dir: %w", err)
	}
	keyDir := home + "/.sops"
	keyFile := keyDir + "/age-key.txt"

	if _, err := os.Stat(keyFile); err == nil {
		slog.Info("🔑 Age key already exists, skipping generation", "path", keyFile)
		return s.updateSopsYAMLFromKey(keyFile)
	}

	slog.Info("🔑 Generating age key", "path", keyFile)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create %s: %w", keyDir, err)
	}

	if err := s.executor.Run(context.Background(), "age-keygen", false, "-o", keyFile); err != nil {
		return fmt.Errorf("age-keygen failed: %w", err)
	}
	slog.Info("🔑 Age key generated")

	return s.updateSopsYAMLFromKey(keyFile)
}

func (s *SopsRunner) updateSopsYAMLFromKey(keyFile string) error {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	var pubKey string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# public key:") {
			pubKey = strings.TrimSpace(strings.TrimPrefix(line, "# public key:"))
			break
		}
	}
	if pubKey == "" {
		return fmt.Errorf("could not find public key in %s", keyFile)
	}

	slog.Info("🔑 Updating .sops.yaml", "publicKey", pubKey)

	sopsData, err := os.ReadFile(".sops.yaml")
	if err != nil {
		return fmt.Errorf("failed to read .sops.yaml: %w", err)
	}

	lines := strings.Split(string(sopsData), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "age") && !strings.HasPrefix(trimmed, "age:") {
			indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
			lines[i] = indent + pubKey
			break
		}
	}

	if err := os.WriteFile(".sops.yaml", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write .sops.yaml: %w", err)
	}

	os.Setenv("SOPS_AGE_KEY_FILE", keyFile)
	slog.Info("✅ .sops.yaml updated and SOPS_AGE_KEY_FILE set")
	return nil
}
