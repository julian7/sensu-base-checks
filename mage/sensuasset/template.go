package sensuasset

import (
	"bytes"
	"html/template"
)

type PackageTemplate struct {
	Arch        string
	ArchiveName string
	Ext         string
	OS          string
	PackageName string
	Version     string
}

func NewPkgTemplate(arch, ext, os, pkgName, version string) *PackageTemplate {
	return &PackageTemplate{
		Arch:        arch,
		Ext:         ext,
		OS:          os,
		PackageName: pkgName,
		Version:     version,
	}
}

func (ts *PackageTemplate) Parse(name, text string) (string, error) {
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
