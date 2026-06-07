package main

import "fmt"

// nava's targets are run through Mage, not this binary.
// Use `mage -l` to list targets, e.g. `mage helm:install` or `mage ko:build`.
func main() {
	fmt.Println("nava is a Mage-driven toolkit. Run `mage -l` to list targets.")
}
