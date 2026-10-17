// Package cimagex exposes package-level convenience functions for CI pipeline
// targets. It wraps pkg/ci.
package cimagex

import (
	cix "github.com/nirantaraai/nava/pkg/ci"
)

// defaultRunner is the package-level runner used by all convenience functions.
var defaultRunner = cix.NewCIRunner()

// LoadConfig loads CI configuration from a YAML file into the default runner.
func LoadConfig(path string) error {
	return defaultRunner.LoadConfig(path)
}

// VerifyGenerated runs a generation command and checks for git diff.
func VerifyGenerated(generateCmd string, generateArgs ...string) error {
	return defaultRunner.VerifyGenerated(generateCmd, generateArgs...)
}

// VerifyGeneratedMage runs `mage <target>` and checks for git drift.
func VerifyGeneratedMage(target string) error {
	return defaultRunner.VerifyGeneratedMage(target)
}

// VerifyModTidy runs `go mod tidy` and checks for git diff on go.mod/go.sum.
func VerifyModTidy() error {
	return defaultRunner.VerifyModTidy()
}

// VerifyNoUntracked checks that no untracked files exist in the given paths.
func VerifyNoUntracked(paths ...string) error {
	return defaultRunner.VerifyNoUntracked(paths...)
}

// GoVet runs go vet with optional build tags.
func GoVet(tags string) error {
	return defaultRunner.GoVet(tags)
}

// Check runs a composite CI gate: vet + mod tidy verification.
func Check(tags string) error {
	return defaultRunner.Check(tags)
}

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	return cix.IsCI()
}

// EnvOrDefault returns the environment variable value or the default.
func EnvOrDefault(key, defaultVal string) string {
	return cix.EnvOrDefault(key, defaultVal)
}
