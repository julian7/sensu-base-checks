package target

import (
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Target struct {
	OS   string
	Arch string
	Ext  string
}

func BuildTarget(os, arch, ext string) *Target {
	return &Target{
		OS:   os,
		Arch: arch,
		Ext:  ext,
	}
}

func (t *Target) WithEnv(version string) map[string]string {
	return map[string]string{
		"GOOS":    t.OS,
		"GOARCH":  t.Arch,
		"VERSION": version,
	}
}

func (t *Target) Compile(version, ldflags, output, pkg string) error {
	vars := t.WithEnv(version)

	err := sh.RunWith(vars, mg.GoCmd(), "build", "-o", output, "-ldflags", ldflags, pkg)
	if err != nil {
		_ = sh.Rm(output)
		return err
	}

	if _, err := exec.LookPath("upx"); err != nil {
		return sh.RunWith(vars, "upx", output)
	}

	return nil
}
