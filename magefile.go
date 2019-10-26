//+build mage

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/julian7/sensu-base-checks/mage/target"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v2"
)

var (
	hasUPX        = false
	buildconfName = "./buildconf.yml"
	buildConfig   = BuildConfig{}
	packageName   = "sensu-base-checks"
	entrypoint    = "./cmd/sensu-base-checks"
	targetDir     = "target"
	ldFlags       = `-s -w -X main.version=$VERSION`
	targets       = map[string]*target.Target{
		"linux":   target.BuildTarget("linux", "amd64", ""),
		"windows": target.BuildTarget("windows", "amd64", ".exe"),
	}
	versionTag                    = "SNAPSHOT"
	assetSpec                     = AssetSpec{}
	defaultExecutableNameTemplate = "{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}{{.Ext}}"
	defaultPackageNameTemplate    = "{{.PackageName}}-{{.OS}}-{{.Arch}}-{{.Version}}.tar.gz"
	defaultReleaseURLTemplate     = "https://github.com/julian7/{{.PackageName}}/releases/download/{{.Version}}/{{.ArchiveName}}"
)

type Buildconf mg.Namespace

type BuildConfig struct {
	PackageNameTemplate    string `yaml:"pkgname,omitempty"`
	ReleaseURLTemplate     string `yaml:"releaseurl,omitempty"`
	ExecutableNameTemplate string `yaml:"execname,omitempty"`
}

type TemplateSpec struct {
	Arch        string
	ArchiveName string
	Ext         string
	OS          string
	PackageName string
	Version     string
}

type AssetSpec struct {
	Type       string       `yaml:"type"`
	APIVersion string       `yaml:"api_version"`
	Metadata   MetadataSpec `yaml:"metadata"`
	Spec       struct {
		Builds []BuildSpec `yaml:"builds"`
	} `yaml:"spec"`
}

type MetadataSpec struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type BuildSpec struct {
	URL     string            `yaml:"url"`
	SHA512  string            `yaml:"sha512"`
	Filters []string          `yaml:"filters"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

func init() {
	buildConfig.ExecutableNameTemplate = defaultExecutableNameTemplate
	buildConfig.PackageNameTemplate = defaultPackageNameTemplate
	buildConfig.ReleaseURLTemplate = defaultReleaseURLTemplate
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
		buildConfig.ExecutableNameTemplate,
		buildConfig.PackageNameTemplate,
		buildConfig.ReleaseURLTemplate,
	)
}

func step(name string) {
	fmt.Printf("-----> %s\n", name)
}

func (ts *TemplateSpec) Parse(name, text string) (string, error) {
	tmpl := template.New(name)
	_, err := tmpl.Parse(text)

	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, ts); err != nil {
		return "", err
	}
	return out.String(), nil

}

func (conf *BuildConfig) tmplSpec(t *target.Target) *TemplateSpec {
	return &TemplateSpec{
		Arch:        t.Arch,
		Ext:         t.Ext,
		OS:          t.OS,
		PackageName: packageName,
		Version:     versionTag,
	}
}

func (conf *BuildConfig) ExecutableName(t *target.Target) (string, error) {
	tmplSpec := conf.tmplSpec(t)
	return tmplSpec.Parse("executable", conf.ExecutableNameTemplate)
}

func (conf *BuildConfig) InternalExecName(t *target.Target) (string, error) {
	tmplSpec := conf.tmplSpec(t)
	return tmplSpec.Parse("executable", "{{.PackageName}}{{.Ext}}")
}

func (conf *BuildConfig) ArchiveName(t *target.Target) (string, error) {
	tmplSpec := conf.tmplSpec(t)
	return tmplSpec.Parse("archive", conf.PackageNameTemplate)
}

func (conf *BuildConfig) URL(t *target.Target) (string, error) {
	archiveName, err := conf.ArchiveName(t)
	if err != nil {
		return "", fmt.Errorf("cannot generate archive name for URL: %w", err)
	}
	tmplSpec := conf.tmplSpec(t)
	tmplSpec.ArchiveName = archiveName
	return tmplSpec.Parse("url", conf.ReleaseURLTemplate)
}

func archive(t *target.Target) error {
	builtexec, err := buildConfig.ExecutableName(t)
	if err != nil {
		return fmt.Errorf("cannot get built executable name: %w", err)
	}
	intexec, err := buildConfig.InternalExecName(t)
	if err != nil {
		return fmt.Errorf("cannot get internal executable name: %w", err)
	}
	archiveFile, err := buildConfig.ArchiveName(t)
	if err != nil {
		return fmt.Errorf("cannot get archive file name: %w", err)
	}
	archiveFile = path.Join(targetDir, archiveFile)
	archive, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("cannot create archive file: %w", err)
	}
	defer archive.Close()

	compressedArchive := gzip.NewWriter(archive)
	defer compressedArchive.Close()

	tw := tar.NewWriter(compressedArchive)
	defer tw.Close()

	if err := writeFileToTar("README.md", "README.md", 0o644, tw); err != nil {
		fmt.Println("No README.md file found. Skipping.")
	}
	if err := writeFileToTar("CHANGELOG.md", "CHANGELOG.md", 0o644, tw); err != nil {
		fmt.Println("No CHANGELOG.md file found. Skipping.")
	}
	if err := writeFileToTar("LICENSE", "LICENSE", 0o644, tw); err != nil {
		if err := writeFileToTar("LICENSE.md", "LICENSE.md", 0o644, tw); err != nil {
			fmt.Println("No LICENSE or LICENSE.md file found. Skipping.")
		}
	}
	if err := writeDirToTar("bin/", 0o755, tw); err != nil {
		sh.Rm(archiveFile)
		return err
	}
	if err := writeFileToTar(
		path.Join(targetDir, builtexec),
		path.Join("bin", intexec),
		0o755,
		tw,
	); err != nil {
		sh.Rm(archiveFile)
		return err
	}
	if err := tw.Close(); err != nil {
		sh.Rm(archiveFile)
		return err
	}
	fmt.Printf("Archive file created: %s\n", archiveFile)
	return nil
}

func writeDirToTar(targetName string, mode int64, tw *tar.Writer) error {
	hdr := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     targetName,
		Mode:     mode,
		ModTime:  time.Now(),
		Format:   tar.FormatUSTAR,
	}
	return tw.WriteHeader(hdr)
}

func writeFileToTar(filename, targetName string, mode int64, tw *tar.Writer) error {
	st, err := os.Stat(filename)
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     targetName,
		Size:     st.Size(),
		Mode:     mode,
		ModTime:  st.ModTime(),
		Format:   tar.FormatUSTAR,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	fileReader, err := os.Open(filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tw, fileReader); err != nil {
		return err
	}
	fileReader.Close()
	return nil
}

func summarize(t *target.Target) error {
	if assetSpec.Spec.Builds == nil {
		assetSpec.Spec.Builds = []BuildSpec{}
	}
	filename, err := buildConfig.ArchiveName(t)
	if err != nil {
		return fmt.Errorf("cannot summarize archive: %w", err)
	}

	url, err := buildConfig.URL(t)
	if err != nil {
		return fmt.Errorf("cannot summarize archive: %w", err)
	}
	hash := sha512.New()
	reader, err := os.Open(path.Join(targetDir, filename))
	if err != nil {
		return fmt.Errorf("cannot open file for calculating SHA512 checksum: %w", err)
	}
	defer reader.Close()
	if _, err := io.Copy(hash, reader); err != nil {
		return fmt.Errorf("cannot read file for calculating SHA512 checksum: %w", err)
	}
	spec := BuildSpec{
		URL:    url,
		SHA512: fmt.Sprintf("%x", hash.Sum(nil)),
		Filters: []string{
			fmt.Sprintf("entity.system.os == '%s'", t.OS),
			fmt.Sprintf("entity.system.arch == '%s'", t.Arch),
		},
	}
	assetSpec.Spec.Builds = append(assetSpec.Spec.Builds, spec)
	return nil
}

func build(name string, t *target.Target) error {
	step(name)
	execname, err := buildConfig.ExecutableName(t)
	if err != nil {
		return fmt.Errorf("cannot build target filename from template: %w", err)
	}
	execfile := path.Join(targetDir, execname)

	err = t.Compile(versionTag, ldFlags, execfile, entrypoint)
	if err != nil {
		return fmt.Errorf("cannot compile executable: %w", err)
	}
	fmt.Printf("Executable created: %s\n", execfile)
	if err := archive(t); err != nil {
		return fmt.Errorf("cannot create archive for executable: %w", err)
	}
	return nil
}

// All builds for all possible targets
func All() error {
	mg.Deps(Buildconf.Read, Target, Version)
	step("all")
	for name, target := range targets {
		if err := build(name, target); err != nil {
			return err
		}
	}
	mg.Deps(Assetfile)
	return nil
}

func Assetfile() error {
	step("assetfile")
	mg.Deps(Buildconf.Read, Target, Version)
	assetSpec.Type = "Asset"
	assetSpec.APIVersion = "core/v2"
	assetSpec.Metadata = MetadataSpec{
		Name:      packageName,
		Namespace: "default",
	}
	for _, target := range targets {
		if err := summarize(target); err != nil {
			return err
		}
	}
	d, err := yaml.Marshal(&assetSpec)
	if err != nil {
		return fmt.Errorf("unable to marshal asset spec: %w", err)
	}
	assetFileName := path.Join(targetDir, fmt.Sprintf("%s-asset-%s.yml", packageName, versionTag))
	assetFile, err := os.Create(assetFileName)
	if err != nil {
		return fmt.Errorf("unable to open asset file to write: %w", err)
	}
	fmt.Fprintf(assetFile, "---\n%s\n", string(d))
	assetFile.Close()
	fmt.Printf("Asset file created: %s\n", assetFileName)
	return nil
}

// Target creates a target directory if not existing
func Target() error {
	st, err := os.Stat(targetDir)
	if err == nil && st.IsDir() {
		return nil
	}
	return os.Mkdir(targetDir, 0o755)
}

func Version() error {
	var err error
	versionTag, err = sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return err
	}
	return nil
}

func Tests() {
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
