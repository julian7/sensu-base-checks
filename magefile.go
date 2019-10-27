//+build mage

package main

import (
	"errors"
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
	config        *Config
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

type Config struct {
	Build        *asset.Config `yaml:"build,omitempty"`
	UploadTarget string        `yaml:"uploadtarget,omitempty"`
}

type Buildconf mg.Namespace

func init() {
	config = &Config{
		Build: asset.NewConfig(
			"{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}.tar.gz",
			"https://github.com/julian7/{{.PackageName}}/releases/download/{{.Version}}/{{.ArchiveName}}",
			"{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}{{.Ext}}",
		),
	}
	config.Build.SetBuildParams(entrypoint, ldFlags, packageName)
	config.Build.SetVersion(versionTag)
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
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return err
	}
	return nil
}

func (c Buildconf) Show() {
	mg.Deps(Buildconf.Read)
	fmt.Printf(
		"Buildconfig:\nexecname: %s\npkgname: %s\nreleaseurl: %s\nuploadtarget: %s\n",
		config.Build.ExecTmpl,
		config.Build.PkgTmpl,
		config.Build.ReleaseURL,
		config.UploadTarget,
	)
}

func step(name string) {
	fmt.Printf("-----> %s\n", name)
}

func assetFile() string {
	return path.Join(targetDir, fmt.Sprintf("%s-asset-%s.yml", packageName, versionTag))
}

// All builds for all possible targets
func All() error {
	mg.Deps(Buildconf.Read, createTargetDir, version)
	step("all")
	for name, target := range targets {
		tar, err := asset.NewTarget(config.Build, target)
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
	assetSpec = config.Build.NewAssetSpec()
	for _, target := range targets {
		t, err := asset.NewTarget(config.Build, target)
		if err != nil {
			return err
		}
		build, err := t.Summarize()
		if err != nil {
			return err
		}
		assetSpec.AddBuild(build)
	}
	assetFilename := assetFile()
	assetfile, err := os.Create(assetFilename)
	if err != nil {
		return fmt.Errorf("unable to open asset file to write: %w", err)
	}
	if _, err := assetSpec.Write(assetfile); err != nil {
		return err
	}
	assetfile.Close()
	fmt.Printf("Asset file created: %s\n", assetFilename)
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
	config.Build.SetVersion(versionTag)
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

// Upload uploads files using SCP
func Upload() error {
	var assets asset.AssetSpec
	mg.Deps(Buildconf.Read, createTargetDir, version)

	if len(config.UploadTarget) == 0 {
		return errors.New("upload target is not specified")
	}
	assetfile := assetFile()

	_, err := os.Stat(assetfile)
	if os.IsNotExist(err) {
		mg.Deps(All)
	}

	step("upload")

	contents, err := ioutil.ReadFile(assetfile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(contents, &assets); err != nil {
		return err
	}
	cmdArgs := []string{assetfile}
	for _, asset := range assets.Spec.Builds {
		cmdArgs = append(cmdArgs, path.Join(targetDir, path.Base(asset.URL)))
	}
	cmdArgs = append(cmdArgs, config.UploadTarget)
	fmt.Printf("Uploading assets version %s\n", versionTag)
	return sh.RunV("scp", cmdArgs...)
}
