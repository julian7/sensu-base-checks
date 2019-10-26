package sensuasset

import (
	"github.com/julian7/sensu-base-checks/mage/target"
)

type Config struct {
	PkgTmpl     string `yaml:"pkgname,omitempty"`
	ReleaseURL  string `yaml:"releaseurl,omitempty"`
	ExecTmpl    string `yaml:"execname,omitempty"`
	buildpkg    string `yaml:"-"`
	ldflags     string `yaml:"-"`
	outdir      string `yaml:"-"`
	packageName string `yaml:"-"`
	versionTag  string `yaml:"-"`
}

func NewConfig(pkgtmpl, relurl, exectmpl string) *Config {
	return &Config{
		PkgTmpl:    pkgtmpl,
		ReleaseURL: relurl,
		ExecTmpl:   exectmpl,
		outdir:     "target",
	}
}

func (conf *Config) SetBuildParams(buildpkg, ldflags, packageName string) {
	conf.buildpkg = buildpkg
	conf.ldflags = ldflags
	conf.packageName = packageName
}

func (conf *Config) SetVersion(version string) {
	conf.versionTag = version
}

func (conf *Config) WithTarget(t *target.Target) (*Target, error) {
	return NewTarget(conf, t)
}
