//go:build mage
// +build mage

package main

import (
	"fmt"
	"os/exec"
)

func BuildWpctl() error {
	buildCmd := exec.Command(
		"go",
		"build",
		"-o",
		"bin/wpctl",
		"cmd/wpctl/main.go",
	)

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed for wpctl with output '%s': %w", output, err)
	}

	fmt.Println("wpctl binary built and available at bin/wpctl")

	return nil
}
