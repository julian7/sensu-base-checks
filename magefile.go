//+build mage

package main

import (
	"fmt"
	"os/exec"

	"github.com/julian7/goshipdone"
	"github.com/julian7/sensulib/sensuasset"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Buildconf mg.Namespace

func step(name string) {
	fmt.Printf("-----> %s\n", name)
}

// All builds for all possible targets
func All() error {
	step("all")
	sensuasset.Register()
	if err := goshipdone.Run(""); err != nil {
		return err
	}
	return nil
}

// Alltests runs all code checks
func Alltests() {
	mg.SerialDeps(Lint, Test, Cover)
}

// Test runs all tests
func Test() error {
	step("test")
	return sh.RunV(mg.GoCmd(), "test", "./...")
}

// Lint tries to run golangci-lint, or golint. If neither of them are available,
// it runs go fmt and go vet instead.
func Lint() error {
	step("lint")
	if _, err := exec.LookPath("golangci-lint"); err == nil {
		return sh.RunV("golangci-lint", "run", "-v", "./...")
	}
	if _, err := exec.LookPath("golint"); err == nil {
		return sh.RunV("golint", "./...")
	}
	if err := sh.RunV(mg.GoCmd(), "fmt", "./..."); err != nil {
		return err
	}
	return sh.RunV(mg.GoCmd(), "vet", "./...")
}

// Cover runs coverity profile
func Cover() error {
	step("cover")
	err := sh.RunV(mg.GoCmd(), "test", "-coverprofile", "sum.cov", "./...")
	if err != nil {
		return err
	}
	return sh.RunV(mg.GoCmd(), "tool", "cover", "-func", "sum.cov")
}
