package sopsmagex

import (
	"fmt"

	"github.com/nirantaraai/nava/pkg/sops"
)

const defaultConfigPath = "k3d.yaml"

// Package-level runner, initialized by LoadConfig.
var defaultRunner *sopsx.SopsRunner

// LoadConfig loads SOPS secret paths from a YAML config file.
func LoadConfig(filepath string) error {
	runner, err := sopsx.NewSopsRunnerFromYAML(filepath)
	if err != nil {
		return err
	}
	defaultRunner = runner
	return nil
}

// LoadDefaultConfig loads from the default "k3d.yaml" path.
func LoadDefaultConfig() error {
	return LoadConfig(defaultConfigPath)
}

func ensureLoaded() error {
	if defaultRunner == nil {
		return fmt.Errorf("sops config not loaded; call LoadConfig() first")
	}
	return nil
}

// Init installs sops+age, generates age key, updates .sops.yaml, and encrypts all secrets.
func Init() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.Init()
}

// Encrypt encrypts all .dec.yaml secret files to .enc.yaml.
func Encrypt() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.EncryptSecrets()
}

// Decrypt decrypts all .enc.yaml secret files to .dec.yaml on disk.
func Decrypt() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.DecryptSecrets()
}

// InstallTools installs sops and age via go install.
func InstallTools() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.InstallTools()
}

// GenerateAgeKey generates an age key and updates .sops.yaml.
func GenerateAgeKey() error {
	if err := ensureLoaded(); err != nil {
		return err
	}
	return defaultRunner.GenerateAgeKey()
}
