package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// FreeBSD pkg manifests use the UCL format
// (https://github.com/vstakhov/libucl). Fortunately, UCL is officially fully
// compatible with JSON so we can just generate JSON data and be done with it.

// See https://github.com/freebsd/pkg/blob/master/libpkg/pkg_manifest.c

type Manifest struct {
	ABI         string              `json:"abi"`
	Comment     string              `json:"comment"`
	Deps        ManifestDeps        `json:"deps,omitempty"`
	Desc        string              `json:"desc"`
	Directories ManifestDirectories `json:"directories,omitempty"`
	Files       ManifestFiles       `json:"files,omitempty"`
	Groups      []string            `json:"groups,omitempty"`
	Licenses    []string            `json:"licenses,omitempty"`
	Maintainer  string              `json:"maintainer,omitempty"`
	Name        string              `json:"name"`
	Origin      string              `json:"origin"`
	Prefix      string              `json:"prefix,omitempty"`
	Scripts     map[string]string   `json:"scripts"`
	Users       []string            `json:"users,omitempty"`
	Version     string              `json:"version"`
	WWW         string              `json:"www,omitempty"`
}

type ManifestDep struct {
	Origin  string `json:"origin"`
	Version string `json:"version"`
}

type ManifestDeps map[string]ManifestDep

type ManifestFile struct {
	Uname string `json:"uname"`
	Gname string `json:"gname"`
	Perm  string `json:"perm"`
	Sum   string `json:"sum"`
}

type ManifestFiles map[string]ManifestFile

type ManifestDirectory struct {
	Uname string `json:"uname"`
	Gname string `json:"gname"`
	Perm  string `json:"perm"`
}

type ManifestDirectories map[string]ManifestDirectory

func NewManifest() *Manifest {
	return &Manifest{
		Files:       make(ManifestFiles),
		Directories: make(ManifestDirectories),
		Scripts:     make(map[string]string),
	}
}

func (m *Manifest) PackageFilename() string {
	return m.Name + "-" + m.Version + ".pkg"
}

func (m *Manifest) WriteFile(filePath string) error {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("cannot encode manifest: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
