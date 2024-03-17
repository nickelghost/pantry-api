//go:build mage
// +build mage

package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

func Test() error {
	if err := sh.RunV(
		"go",
		"test",
		"-coverprofile=coverage.out",
		"./...",
	); err != nil {
		return err
	}

	if err := sh.RunV(
		"go",
		"tool",
		"cover",
		"-html=coverage.out",
		"-o=coverage.html",
	); err != nil {
		return err
	}

	return os.Remove("coverage.out")
}
