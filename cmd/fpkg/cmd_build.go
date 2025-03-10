package main

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"slices"
	"sort"
	"strconv"
	"time"
	"unicode"

	"go.n16f.net/program"
)

func cmdBuild(p *program.Program) {
	outputPath := p.OptionValue("output")

	dirPath := p.ArgumentValue("directory")
	if dirPath == "" {
		dirPath = "."
	}

	configPath := p.OptionValue("config")
	config := DefaultGenerationConfig()
	if err := config.LoadFile(configPath); err != nil {
		p.Fatal("cannot load configuration file from %q: %v", configPath, err)
	}

	if p.IsOptionSet("version") {
		config.Version = p.OptionValue("version")
	}

	if config.Version == "" {
		p.Fatal("missing or empty version")
	}

	manifest, err := generateManifest(config, dirPath)
	if err != nil {
		p.Fatal("cannot generate manifest: %v", err)
	}

	archivePath := path.Join(outputPath, manifest.PackageFilename())

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	archive, err := os.OpenFile(archivePath, flags, 0644)
	if err != nil {
		p.Fatal("cannot open %q: %v", archivePath, err)
	}

	if err := createArchive(config, dirPath, manifest, archive); err != nil {
		if removeErr := os.Remove(archivePath); removeErr != nil {
			p.Error("cannot delete %q: %v", archivePath, removeErr)
		}

		p.Fatal("cannot create archive: %v", err)
	}

	fmt.Printf("%s\n", archivePath)
}

func generateManifest(config *GenerationConfig, dirPath string) (*Manifest, error) {
	now := time.Now()

	version, err := Version(buildId)
	if err != nil {
		return nil, err
	}

	m := NewManifest()

	m.Annotations = map[string]string{
		"built_by":        "fpkg-" + version,
		"build_timestamp": now.UTC().Format(time.RFC3339),
	}

	m.Name = config.Name
	m.Version = config.Version
	m.Comment = config.ShortDescription
	m.Desc = config.LongDescription
	m.WWW = config.WebsiteURI
	m.Maintainer = config.Maintainer
	m.Licenses = slices.Clone(config.Licenses)
	m.Categories = slices.Clone(config.Categories)

	if abi := config.ABI; abi != "" {
		m.ABI = abi
	} else {
		m.ABI = "*"
	}

	if longDesc := config.LongDescription; longDesc != "" {
		m.Desc = longDesc
	} else {
		desc := []rune(m.Comment)
		desc[0] = unicode.ToUpper(desc[0])
		m.Desc = string(desc) + "."
	}

	if origin := config.Origin; origin != "" {
		m.Origin = origin
	} else {
		m.Origin = "misc/" + config.Name
	}

	m.Deps = make(ManifestDeps, len(config.Dependencies))
	for _, dep := range config.Dependencies {
		m.Deps[dep.Name] = ManifestDep{
			Origin:  dep.Name,
			Version: dep.Version,
		}
	}

	m.Users = make([]string, len(config.Users))
	for i, user := range config.Users {
		m.Users[i] = user.Name
	}

	m.Groups = make([]string, len(config.Groups))
	for i, group := range config.Groups {
		m.Groups[i] = group.Name
	}

	m.Prefix = "/"

	for _, dir := range config.Directories {
		var mdir ManifestDirectory

		if dir.Owner != "" {
			mdir.Uname = dir.Owner
		} else {
			mdir.Uname = config.FileOwner
		}

		if dir.Group != "" {
			mdir.Gname = dir.Group
		} else {
			mdir.Gname = config.FileGroup
		}

		if dir.Mode != "" {
			mdir.Perm = dir.Mode
		} else {
			mdir.Perm = "755"
		}

		m.Directories[dir.Path] = mdir
	}

	err = WalkDir(dirPath, func(relPath string, info fs.FileInfo) error {
		fullPath := path.Join(dirPath, relPath)

		if !info.Mode().IsRegular() {
			return nil
		}

		checksum, err := FileSHA256Checksum(fullPath)
		if err != nil {
			return fmt.Errorf("cannot compute checksum of %q: %w",
				fullPath, err)
		}

		fileCfg, hasFileCfg := config.FindFile(relPath)

		var permString string
		if hasFileCfg && fileCfg.Mode != "" {
			permString = fileCfg.Mode
		} else {
			permString = strconv.FormatInt(int64(info.Mode().Perm()), 8)
		}

		var uname string
		if hasFileCfg && fileCfg.Owner != "" {
			uname = fileCfg.Owner
		} else {
			uname = config.FileOwner
		}

		var gname string
		if hasFileCfg && fileCfg.Group != "" {
			gname = fileCfg.Group
		} else {
			gname = config.FileGroup
		}

		if info.IsDir() {
			m.Directories[relPath] = ManifestDirectory{
				Uname: uname,
				Gname: gname,
				Perm:  permString,
			}
		} else {
			m.Files[relPath] = ManifestFile{
				Uname: uname,
				Gname: gname,
				Perm:  permString,
				Sum:   hex.EncodeToString(checksum),
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	preInstallData, err := generatePreInstall(config)
	if err != nil {
		return nil, fmt.Errorf("cannot generate pre-install script: %w", err)
	}

	m.Scripts["pre-install"] = string(preInstallData)

	return m, nil
}

func createArchive(config *GenerationConfig, dirPath string, manifest *Manifest, archive io.Writer) error {
	now := time.Now().UTC()

	w := tar.NewWriter(archive)

	addFile := func(name string, mode int64, owner, group string, data []byte) error {
		header := tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(data)),
			Mode:     mode,
			ModTime:  now,
			Uname:    owner,
			Gname:    group,
		}

		if data == nil {
			header.Typeflag = tar.TypeDir
		}

		if owner := config.FileOwner; owner != "" {
			header.Uname = owner
		}

		if group := config.FileGroup; group != "" {
			header.Gname = group
		}

		if err := w.WriteHeader(&header); err != nil {
			return fmt.Errorf("cannot write header: %w", err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("cannot write data: %w", err)
		}

		return nil
	}

	// Manifest
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("cannot encode manifest: %w", err)
	}

	err = addFile("+MANIFEST", 0644, config.FileOwner, config.FileGroup,
		manifestData)
	if err != nil {
		return fmt.Errorf("cannot add manifest: %w", err)
	}

	// Files
	relPaths := make([]string, 0, len(manifest.Files))
	for relPath := range manifest.Files {
		relPaths = append(relPaths, relPath)
	}
	sort.Strings(relPaths)

	for _, relPath := range relPaths {
		mfile := manifest.Files[relPath]
		filePath := path.Join(dirPath, relPath)

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read %q: %w", filePath, err)
		}

		perm, err := strconv.ParseInt(mfile.Perm, 8, 64)
		if err != nil {
			return fmt.Errorf("cannot parse permission string %q: %w",
				mfile.Perm, err)
		}

		err = addFile(relPath, perm, mfile.Uname, mfile.Gname, data)
		if err != nil {
			return fmt.Errorf("cannot add %q: %w", filePath, err)
		}
	}

	// Directories
	relPaths = make([]string, 0, len(manifest.Directories))
	for relPath := range manifest.Directories {
		relPaths = append(relPaths, relPath)
	}
	sort.Strings(relPaths)

	for _, relPath := range relPaths {
		mdir := manifest.Directories[relPath]
		filePath := path.Join(dirPath, relPath)

		perm, err := strconv.ParseInt(mdir.Perm, 8, 64)
		if err != nil {
			return fmt.Errorf("cannot parse permission string %q: %w",
				mdir.Perm, err)
		}

		err = addFile(relPath, perm, mdir.Uname, mdir.Gname, nil)
		if err != nil {
			return fmt.Errorf("cannot add directory %q: %w", filePath, err)
		}
	}

	// Finalize the archive
	if err := w.Close(); err != nil {
		return fmt.Errorf("cannot close archive: %w", err)
	}

	return nil
}

func generatePreInstall(config *GenerationConfig) ([]byte, error) {
	if len(config.Groups) == 0 && len(config.Users) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer

	buf.WriteString(`
if [ -n "$PKG_ROOTDIR" ] && [ "$PKG_ROOTDIR" != "/" ]; then
  PW="/usr/sbin/pw -R $PKG_ROOTDIR"
else
  PW=/usr/sbin/pw
fi
`)

	if len(config.Groups) > 0 {
		buf.WriteString(`
echo "===> Creating groups."
`)
		for _, group := range config.Groups {
			fmt.Fprintf(&buf, `
if ! $PW groupshow '%s' >/dev/null 2>&1; then
  echo "Creating group '%s' with gid %d."
  $PW groupadd '%s' -g %d
else
  echo "Using existing group '%s'."
fi
`,
				group.Name,
				group.Name, group.GID,
				group.Name, group.GID,
				group.Name)
		}
	}

	if len(config.Users) > 0 {
		buf.WriteString(`
echo "===> Creating users."
`)
		for _, user := range config.Users {
			fmt.Fprintf(&buf, `
if ! $PW usershow '%s' >/dev/null 2>&1; then
  echo "Creating user '%s' with uid %d."
  $PW useradd '%s' -u %d -g %s -c '%s' \
                   -d /nonexistent -s /usr/sbin/nologin
else
  echo "Using existing user '%s'."
fi
`,
				user.Name,
				user.Name, user.UID,
				user.Name, user.UID, user.Group, user.Name,
				user.Name)
		}
	}

	return buf.Bytes(), nil
}
