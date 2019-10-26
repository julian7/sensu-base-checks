package sensuasset

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/julian7/sensu-base-checks/mage/target"
	"github.com/magefile/mage/sh"
)

type Target struct {
	*Config
	target       *target.Target
	executable   string
	packagedexec string
	archivefile  string
	url          string
}

func NewTarget(conf *Config, bldTarget *target.Target) (*Target, error) {
	var err error

	t := &Target{Config: conf, target: bldTarget}
	tmplSpec := NewPkgTemplate(
		bldTarget.Arch,
		bldTarget.Ext,
		bldTarget.OS,
		conf.packageName,
		conf.versionTag,
	)

	t.executable, err = tmplSpec.Parse("executable", t.ExecTmpl)
	if err != nil {
		return nil, fmt.Errorf("cannot build new target: %w", err)
	}

	t.packagedexec, err = tmplSpec.Parse("pkgexec", "{{.PackageName}}{{.Ext}}")
	if err != nil {
		return nil, fmt.Errorf("cannot build new target: %w", err)
	}

	t.archivefile, err = tmplSpec.Parse("archive", t.PkgTmpl)
	if err != nil {
		return nil, fmt.Errorf("cannot build new target: %w", err)
	}

	t.url, err = tmplSpec.Parse("url", t.ReleaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot build new target: %w", err)
	}

	return t, nil
}

func (t *Target) Archive() (string, error) {
	archiveFile := path.Join(t.outdir, t.archivefile)

	archive, err := os.Create(archiveFile)
	if err != nil {
		return "", fmt.Errorf("cannot create archive file: %w", err)
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
		_ = sh.Rm(archiveFile)
		return "", err
	}

	if err := writeFileToTar(
		path.Join(t.outdir, t.executable),
		path.Join("bin", t.packagedexec),
		0o755,
		tw,
	); err != nil {
		_ = sh.Rm(archiveFile)
		return "", err
	}

	if err := tw.Close(); err != nil {
		_ = sh.Rm(archiveFile)
		return "", err
	}

	return archiveFile, nil
}

func (t *Target) Compile() (string, error) {
	outfile := path.Join(t.Config.outdir, t.executable)
	err := t.target.Compile(
		t.Config.versionTag,
		t.Config.ldflags,
		outfile,
		t.Config.buildpkg,
	)

	return outfile, err
}

func (t *Target) Summarize() (*BuildSpec, error) {
	archiveFile := path.Join(t.outdir, t.archivefile)

	reader, err := os.Open(archiveFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for calculating SHA512 checksum: %w", err)
	}
	defer reader.Close()

	hash := sha512.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return nil, fmt.Errorf("cannot read file for calculating SHA512 checksum: %w", err)
	}

	build := &BuildSpec{
		URL:    t.url,
		SHA512: fmt.Sprintf("%x", hash.Sum(nil)),
		Filters: []string{
			fmt.Sprintf("entity.system.os == '%s'", t.target.OS),
			fmt.Sprintf("entity.system.arch == '%s'", t.target.Arch),
		},
	}

	return build, nil
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
