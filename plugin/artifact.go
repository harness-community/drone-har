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

// Wire format written to PLUGIN_ARTIFACT_FILE (CI-23036). lite-engine forwards
// this file to CI manager, which surfaces the entries in step outputs, the
// artifacts tab and the artifactDetails MongoDB collection.
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

// writeArtifactFile writes the uploaded file metadata to the path in
// PLUGIN_ARTIFACT_FILE so that CI manager can pick it up and store it.
// All errors are soft-failed with a warning so the step is never failed.
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

// computeFileSHA256 returns "sha256:<hex>" for the file at path.
// The digest is always computed from the local source file (not provider
// output). Returns an empty string on error (soft-fail).
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
		filePath := filePathInPackage(config, f, sourceRoot)
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

// escapeFilePath percent-encodes each segment of a package-relative file path so
// that names containing spaces or URL-reserved characters yield a valid URL.
// Separators are preserved; filePath itself is still reported unescaped.
func escapeFilePath(filePath string) string {
	segments := strings.Split(filePath, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func filePathInPackage(config packages.Config, file, sourceRoot string) string {
	if config.Filename != "" && sourceRoot == "" {
		return strings.TrimLeft(filepath.ToSlash(config.Filename), "/")
	}
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
