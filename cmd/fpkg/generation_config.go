package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type GenerationConfig struct {
	Name             string                       `yaml:"name"`
	Version          string                       `yaml:"version,omitempty"`
	ShortDescription string                       `yaml:"short_description,omitempty"`
	LongDescription  string                       `yaml:"long_description,omitempty"`
	WebsiteURI       string                       `yaml:"website_uri"`
	Maintainer       string                       `yaml:"maintainer"`
	Licenses         []string                     `yaml:"licenses,omitempty"`
	Categories       []string                     `yaml:"categories,omitempty"`
	Origin           string                       `yaml:"origin,omitempty"`
	ABI              string                       `yaml:"abi,omitempty"`
	Dependencies     []GenerationConfigDependency `yaml:"dependencies,omitempty"`
	Users            []GenerationConfigUser       `yaml:"users,omitempty"`
	Groups           []GenerationConfigGroup      `yaml:"groups,omitempty"`
	FileOwner        string                       `yaml:"file_owner,omitempty"`
	FileGroup        string                       `yaml:"file_group,omitempty"`
	Files            []GenerationConfigFile       `yaml:"files,omitempty"`
	Directories      []GenerationConfigDirectory  `yaml:"directories,omitempty"`
}

type GenerationConfigDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type GenerationConfigUser struct {
	Name  string `yaml:"name"`
	UID   uint   `yaml:"uid"`
	Group string `yaml:"group"`
}

type GenerationConfigGroup struct {
	Name string `yaml:"name"`
	GID  uint   `yaml:"gid"`
}

type GenerationConfigFile struct {
	Path             string `yaml:"path,omitempty"`
	PathRegexpString string `yaml:"path_regexp,omitempty"`
	PathRegexp       *regexp.Regexp
	Mode             string `yaml:"mode,omitempty"`
	Owner            string `yaml:"owner,omitempty"`
	Group            string `yaml:"group,omitempty"`
}

type GenerationConfigDirectory struct {
	Path  string `yaml:"path,omitempty"`
	Mode  string `yaml:"mode,omitempty"`
	Owner string `yaml:"owner,omitempty"`
	Group string `yaml:"group,omitempty"`
}

func DefaultGenerationConfig() *GenerationConfig {
	return &GenerationConfig{
		FileOwner: "root",
		FileGroup: "wheel",
	}
}

func (pc *GenerationConfig) UnmarshalYAML(value *yaml.Node) error {
	type GenerationConfig2 GenerationConfig
	c := GenerationConfig2(*pc)

	if err := value.Decode(&c); err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("missing or empty package name")
	}

	if c.ShortDescription == "" {
		return fmt.Errorf("missing or empty short description")
	}

	if c.WebsiteURI == "" {
		return fmt.Errorf("missing or empty website URI")
	}

	if c.Maintainer == "" {
		return fmt.Errorf("missing or empty maintainer")
	}

	*pc = GenerationConfig(c)
	return nil
}

func (pc *GenerationConfigUser) UnmarshalYAML(value *yaml.Node) error {
	type GenerationConfigUser2 GenerationConfigUser
	c := GenerationConfigUser2(*pc)

	if err := value.Decode(&c); err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("missing or empty user name")
	}

	if c.UID == 0 {
		return fmt.Errorf("missing or zero uid")
	}

	if c.Group == "" {
		return fmt.Errorf("missing or empty user group")
	}

	*pc = GenerationConfigUser(c)
	return nil
}

func (pc *GenerationConfigGroup) UnmarshalYAML(value *yaml.Node) error {
	type GenerationConfigGroup2 GenerationConfigGroup
	c := GenerationConfigGroup2(*pc)

	if err := value.Decode(&c); err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("missing or empty group name")
	}

	if c.GID == 0 {
		return fmt.Errorf("missing or zero gid")
	}

	*pc = GenerationConfigGroup(c)
	return nil
}

func (pc *GenerationConfigFile) UnmarshalYAML(value *yaml.Node) error {
	type GenerationConfigFile2 GenerationConfigFile
	c := GenerationConfigFile2(*pc)

	if err := value.Decode(&c); err != nil {
		return err
	}

	if c.Path == "" && c.PathRegexpString == "" {
		return fmt.Errorf("missing or empty file path or file path regexp")
	}

	if c.Path != "" && c.PathRegexpString != "" {
		return fmt.Errorf("cannot set both file path and file path regexp")
	}

	if s := c.PathRegexpString; s != "" {
		re, err := regexp.Compile(s)
		if err != nil {
			return fmt.Errorf("invalid regexp %q: %w", s, err)
		}

		c.PathRegexp = re
	}

	*pc = GenerationConfigFile(c)
	return nil
}

func (c *GenerationConfig) LoadFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&c); err != nil {
		return fmt.Errorf("cannot decode configuration: %w", err)
	}

	return nil
}

func (c *GenerationConfig) FindFile(filePath string) (GenerationConfigFile, bool) {
	for _, file := range c.Files {
		switch {
		case file.Path != "":
			if file.Path == filePath {
				return file, true
			}

		case file.PathRegexp != nil:
			if file.PathRegexp.MatchString(filePath) {
				return file, true
			}
		}
	}

	return GenerationConfigFile{}, false
}
