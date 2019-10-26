package sensuasset

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type AssetSpec struct {
	Type       string       `yaml:"type"`
	APIVersion string       `yaml:"api_version"`
	Metadata   MetadataSpec `yaml:"metadata"`
	Spec       struct {
		Builds []*BuildSpec `yaml:"builds"`
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

func (conf *Config) NewAssetSpec() *AssetSpec {
	spec := &AssetSpec{
		Type:       "Asset",
		APIVersion: "core/v2",
		Metadata: MetadataSpec{
			Name:      conf.packageName,
			Namespace: "default",
		},
	}

	return spec
}

func (a *AssetSpec) AddBuild(build *BuildSpec) {
	if a.Spec.Builds == nil {
		a.Spec.Builds = []*BuildSpec{}
	}

	a.Spec.Builds = append(a.Spec.Builds, build)
}

func (a *AssetSpec) Write(wr io.Writer) (int, error) {
	d, err := yaml.Marshal(a)
	if err != nil {
		return 0, fmt.Errorf("unable to marshal asset spec: %w", err)
	}

	return fmt.Fprintf(wr, "---\n%s\n", string(d))
}
