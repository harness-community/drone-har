// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/harness/drone-har/plugin/packages"
	"github.com/sirupsen/logrus"
)

type artifactFile struct {
	Kind string           `json:"kind"`
	Data artifactFileData `json:"data"`
}

type artifactFileData struct {
	FileArtifacts []fileArtifactEntry `json:"fileArtifacts"`
}

type fileArtifactEntry struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	FilePath string `json:"filePath,omitempty"`
	Digest   string `json:"digest,omitempty"`
	// HAR-specific fields (null for S3/GCS/Artifactory in other plugins)
	Registry    string `json:"registry,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	Version     string `json:"version,omitempty"`
	PackageType string `json:"packageType,omitempty"`
}

func writeArtifactFile(entries []fileArtifactEntry) {
	path := os.Getenv("PLUGIN_ARTIFACT_FILE")
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		logrus.Warnf("artifact: failed to create directory %s: %v", filepath.Dir(path), err)
		return
	}
	payload := artifactFile{
		Kind: "fileUpload/v1",
		Data: artifactFileData{FileArtifacts: entries},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logrus.Warnf("artifact: failed to marshal metadata: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		logrus.Warnf("artifact: failed to write file %s: %v", path, err)
	}
}

func computeFileSHA256(path string) string {
	f, err := os.Open(path)
	if err != nil {
		logrus.Warnf("artifact: failed to open %s for SHA256 computation: %v", path, err)
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		logrus.Warnf("artifact: failed to compute SHA256 for %s: %v", path, err)
		return ""
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

func collectArtifactEntries(config packages.Config, packageType string) []fileArtifactEntry {
	if !strings.EqualFold(packageType, string(packages.Generic)) {
		logrus.Printf("artifact: skipping artifact metadata publishing for package type %s (only GENERIC is supported)", packageType)
		return nil
	}
	if config.Registry == "" || config.Name == "" || config.Source == "" || config.PkgURL == "" || config.Account == "" {
		logrus.Warnf("artifact: skipping artifact metadata publishing: registry, name, source, pkg-url and account are required")
		return nil
	}

	// Mirror GenericHandler.Push: the version defaults to "v1" when unset.
	version := config.Version
	if version == "" {
		version = "v1"
	}

	files, sourceRoot := resolveUploadFiles(config.Source)
	if len(files) == 0 {
		logrus.Warnf("artifact: no files found for source %s", config.Source)
		return nil
	}

	pkgHost := strings.TrimRight(config.PkgURL, "/")
	var entries []fileArtifactEntry
	for _, f := range files {
		filePath := filePathInPackage(f, sourceRoot)
		entries = append(entries, fileArtifactEntry{
			Name:        filepath.Base(f),
			URL:         harFileURL(pkgHost, config.Account, config.Registry, config.Name, version, filePath),
			FilePath:    filePath,
			Digest:      computeFileSHA256(f),
			Registry:    config.Registry,
			PackageName: config.Name,
			Version:     version,
			PackageType: strings.ToUpper(packageType),
		})
	}
	return entries
}

func harFileURL(pkgHost, account, registry, packageName, version, filePath string) string {
	return strings.Join([]string{pkgHost, "pkg", account, registry, "files", packageName, version, escapeFilePath(filePath)}, "/")
}

func escapeFilePath(filePath string) string {
	segments := strings.Split(filePath, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

// There is no rename case — hc has
// no --filename flag, and a push with PLUGIN_FILENAME set fails at hc before
// metadata is written.
func filePathInPackage(file, sourceRoot string) string {
	if sourceRoot != "" {
		if rel, err := filepath.Rel(sourceRoot, file); err == nil && rel != "" && rel != "." {
			return filepath.ToSlash(filepath.Join(filepath.Base(sourceRoot), rel))
		}
	}
	return filepath.Base(file)
}

func resolveUploadFiles(source string) (files []string, sourceRoot string) {
	root := strings.TrimRight(source, "/")
	info, err := os.Stat(root)
	if err != nil {
		logrus.Warnf("artifact: failed to stat source %s: %v", source, err)
		return nil, ""
	}
	if !info.IsDir() {
		return []string{source}, ""
	}
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip hidden directories (like .git), but never the root itself.
			if path != root && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip hidden files.
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, root
}
