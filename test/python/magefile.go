//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/nirantaraai/nava/mage/python"
)

// PY namespace for PY operations
type PY mg.Namespace

func loadConfig() error {
	return pythonmagex.LoadConfig("pythonmagex.yaml")
}

// Setup verifies toolchain and fetches dependencies
func (PY) Setup() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return pythonmagex.Setup()
}

// Run runs the PY project
func (PY) Run() error {
	if err := loadConfig(); err != nil {
		return err
	}
	return pythonmagex.RunService()
}
