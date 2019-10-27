//+build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/julian7/sensulib/mage/asset"
	"github.com/julian7/sensulib/mage/target"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v2"
)

var (
	buildConfig   *asset.Config
	assetSpec     *asset.AssetSpec
	hasUPX        = false
	buildconfName = "./buildconf.yml"
	packageName   = "sensu-base-checks"
	entrypoint    = "./cmd/sensu-base-checks"
	targetDir     = "target"
	ldFlags       = `-s -w -X main.version=$VERSION`
	targets       = map[string]*target.Target{
		"linux":   target.BuildTarget("linux", "amd64", ""),
		"windows": target.BuildTarget("windows", "amd64", ".exe"),
	}
	versionTag = "SNAPSHOT"
)

type Buildconf mg.Namespace

func init() {
	buildConfig = asset.NewConfig(
		"{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}.tar.gz",
		"https://github.com/julian7/{{.PackageName}}/releases/download/{{.Version}}/{{.ArchiveName}}",
		"{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}{{.Ext}}",
	)
	buildConfig.SetBuildParams(entrypoint, ldFlags, packageName)
	buildConfig.SetVersion(versionTag)
}

func (c Buildconf) Read() error {
	_, err := os.Stat(buildconfName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	contents, err := ioutil.ReadFile(buildconfName)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(contents, &buildConfig); err != nil {
		return err
	}
	return nil
}

func (c Buildconf) Show() {
	mg.Deps(Buildconf.Read)
	fmt.Printf(
		"Buildconfig:\nexecname: %s\npkgname: %s\nreleaseurl: %s\n",
		buildConfig.ExecTmpl,
		buildConfig.PkgTmpl,
		buildConfig.ReleaseURL,
	)
}

func step(name string) {
	fmt.Printf("-----> %s\n", name)
}

// All builds for all possible targets
func All() error {
	mg.Deps(Buildconf.Read, createTargetDir, version)
	step("all")
	for name, target := range targets {
		tar, err := asset.NewTarget(buildConfig, target)
		if err != nil {
			return err
		}
		step(name)
		outfile, err := tar.Compile()
		if err != nil {
			return err
		}
		fmt.Printf("successfuly built %s\n", outfile)
		outfile, err = tar.Archive()
		if err != nil {
			return err
		}
		fmt.Printf("successfuly archived %s\n", outfile)
	}
	mg.Deps(Assetfile)
	return nil
}

// Assetfile generates a new asset yaml file based on already built archives
func Assetfile() error {
	step("assetfile")
	mg.Deps(Buildconf.Read, createTargetDir, version)
	assetSpec = buildConfig.NewAssetSpec()
	for _, target := range targets {
		t, err := asset.NewTarget(buildConfig, target)
		if err != nil {
			return err
		}
		build, err := t.Summarize()
		if err != nil {
			return err
		}
		assetSpec.AddBuild(build)
	}
	assetFileName := path.Join(targetDir, fmt.Sprintf("%s-asset-%s.yml", packageName, versionTag))
	assetFile, err := os.Create(assetFileName)
	if err != nil {
		return fmt.Errorf("unable to open asset file to write: %w", err)
	}
	if _, err := assetSpec.Write(assetFile); err != nil {
		return err
	}
	assetFile.Close()
	fmt.Printf("Asset file created: %s\n", assetFileName)
	return nil
}

func createTargetDir() error {
	st, err := os.Stat(targetDir)
	if err == nil && st.IsDir() {
		return nil
	}
	return os.Mkdir(targetDir, 0o755)
}

func version() error {
	var err error
	versionTag, err = sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return err
	}
	buildConfig.SetVersion(versionTag)
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
