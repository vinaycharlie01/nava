//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/nirantaraai/nava/mage/rust"
)

// Rust namespace for Rust operations
type Rust mg.Namespace

func loadConfig() error {
	return rustmagex.LoadConfig("rustmagex.yaml")
}

// Setup verifies toolchain and fetches dependencies
func (Rust) Setup() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Setup()
}

// Build builds the Rust project
func (Rust) Build() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Build()
}

// Run runs the Rust project
func (Rust) Run() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Run()
}

// Test runs Rust tests
func (Rust) Test() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Test()
}

// Lint runs clippy
func (Rust) Lint() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Lint()
}

// Format checks code formatting
func (Rust) Format() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Format()
}

// Clean cleans build artifacts
func (Rust) Clean() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.Clean()
}

// GenAll runs all configured genCommands
func (Rust) GenAll() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return rustmagex.RunAllGenCommands()
}

// Gen runs one configured genCommand by name
func (Rust) Gen(name string) error {
	if err := loadConfig(); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("please provide a command name, e.g. mage rust:gen check")
	}
	return rustmagex.RunGenCommand(name)
}
