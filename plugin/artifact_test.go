// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/harness/drone-har/plugin/packages"
)

// ---------------------------------------------------------------------------
// writeArtifactFile
// ---------------------------------------------------------------------------

func TestWriteArtifactFile_WritesCorrectJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.json")
	t.Setenv("PLUGIN_ARTIFACT_FILE", path)

	entries := []fileArtifactEntry{
		{
			Name:        "app.jar",
			URL:         "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/1.0.0/app.jar",
			FilePath:    "app.jar",
			Digest:      "sha256:abc123",
			Registry:    "my-registry",
			PackageName: "my-package",
			Version:     "1.0.0",
			PackageType: "GENERIC",
		},
	}
	writeArtifactFile(entries)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("artifact file not written: %v", err)
	}
	var got artifactFile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Kind != "fileUpload/v1" {
		t.Errorf("kind: want fileUpload/v1, got %q", got.Kind)
	}
	if len(got.Data.FileArtifacts) != 1 {
		t.Fatalf("fileArtifacts length: want 1, got %d", len(got.Data.FileArtifacts))
	}
	e := got.Data.FileArtifacts[0]
	if e.Name != "app.jar" {
		t.Errorf("name: want app.jar, got %q", e.Name)
	}
	if e.Digest != "sha256:abc123" {
		t.Errorf("digest: want sha256:abc123, got %q", e.Digest)
	}
	if e.Registry != "my-registry" {
		t.Errorf("registry: want my-registry, got %q", e.Registry)
	}
	if e.PackageName != "my-package" {
		t.Errorf("packageName: want my-package, got %q", e.PackageName)
	}
	if e.Version != "1.0.0" {
		t.Errorf("version: want 1.0.0, got %q", e.Version)
	}
	if e.PackageType != "GENERIC" {
		t.Errorf("packageType: want GENERIC, got %q", e.PackageType)
	}
}

func TestWriteArtifactFile_NoOpWhenEnvUnset(t *testing.T) {
	t.Setenv("PLUGIN_ARTIFACT_FILE", "")
	// Should not panic or create any file.
	writeArtifactFile([]fileArtifactEntry{{Name: "x", URL: "y"}})
}

func TestWriteArtifactFile_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "artifact.json")
	t.Setenv("PLUGIN_ARTIFACT_FILE", path)

	writeArtifactFile([]fileArtifactEntry{{Name: "f", URL: "u"}})

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s, got: %v", path, err)
	}
}

// ---------------------------------------------------------------------------
// computeFileSHA256
// ---------------------------------------------------------------------------

func TestComputeFileSHA256_KnownHash(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(f, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	// echo -n "hello" | sha256sum → 2cf24dba...
	const want = "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	got := computeFileSHA256(f)
	if got != want {
		t.Errorf("SHA256: want %q, got %q", want, got)
	}
}

func TestComputeFileSHA256_ReturnsEmptyOnMissingFile(t *testing.T) {
	got := computeFileSHA256("/nonexistent/path/file.txt")
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// harFileURL
// ---------------------------------------------------------------------------

func TestHarFileURL(t *testing.T) {
	got := harFileURL("https://pkg.harness.io", "acc123", "my-registry", "my-package", "1.0.0", "builds/app.jar")
	want := "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/1.0.0/builds/app.jar"
	if got != want {
		t.Errorf("url: want %q, got %q", want, got)
	}
}

func TestHarFileURL_VanityHost(t *testing.T) {
	// Vanity URLs (self-managed / subdomain) must be used as-is.
	got := harFileURL("https://acme.harness.io", "acc123", "reg", "pkg", "v1", "f.txt")
	want := "https://acme.harness.io/pkg/acc123/reg/files/pkg/v1/f.txt"
	if got != want {
		t.Errorf("url: want %q, got %q", want, got)
	}
}

func TestHarFileURL_EscapesReservedCharacters(t *testing.T) {
	cases := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "space",
			filePath: "my builds/app v2.jar",
			want:     "https://pkg.harness.io/pkg/acc/reg/files/pkg/v1/my%20builds/app%20v2.jar",
		},
		{
			name:     "hash truncates url when unescaped",
			filePath: "builds/app#1.jar",
			want:     "https://pkg.harness.io/pkg/acc/reg/files/pkg/v1/builds/app%231.jar",
		},
		{
			name:     "question mark and percent",
			filePath: "why?/100%.txt",
			want:     "https://pkg.harness.io/pkg/acc/reg/files/pkg/v1/why%3F/100%25.txt",
		},
		{
			name:     "plain names are unchanged",
			filePath: "builds/app-1.0.0_final.jar",
			want:     "https://pkg.harness.io/pkg/acc/reg/files/pkg/v1/builds/app-1.0.0_final.jar",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := harFileURL("https://pkg.harness.io", "acc", "reg", "pkg", "v1", tc.filePath)
			if got != tc.want {
				t.Errorf("url: want %q, got %q", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// filePathInPackage
// ---------------------------------------------------------------------------

func TestFilePathInPackage(t *testing.T) {
	cases := []struct {
		name       string
		file       string
		sourceRoot string
		want       string
	}{
		{
			name:       "single file uses basename",
			file:       "/tmp/build/app.jar",
			sourceRoot: "",
			want:       "app.jar",
		},
		{
			name:       "directory upload prefixes source dir base name",
			file:       "/tmp/build/sub/deep.jar",
			sourceRoot: "/tmp/build",
			want:       "build/sub/deep.jar",
		},
		{
			name:       "directory upload with nested path",
			file:       "/tmp/build/a/b/c/deep.jar",
			sourceRoot: "/tmp/build",
			want:       "build/a/b/c/deep.jar",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := filePathInPackage(tc.file, tc.sourceRoot); got != tc.want {
				t.Errorf("filePath: want %q, got %q", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// resolveUploadFiles
// ---------------------------------------------------------------------------

func TestResolveUploadFiles_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "only.jar")
	writeFile(t, path, "data")

	files, root := resolveUploadFiles(path)
	if len(files) != 1 {
		t.Fatalf("want 1, got %d", len(files))
	}
	if root != "" {
		t.Errorf("sourceRoot should be empty for a single file, got %q", root)
	}
}

func TestResolveUploadFiles_Directory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "a")
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "b.txt"), "b")
	// Hidden files and directories are skipped.
	writeFile(t, filepath.Join(dir, ".hidden"), "h")
	hiddenDir := filepath.Join(dir, ".hidden-dir")
	if err := os.MkdirAll(hiddenDir, 0700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(hiddenDir, "nested.txt"), "g")

	files, root := resolveUploadFiles(dir)
	if len(files) != 2 {
		t.Errorf("want 2 files (hidden skipped), got %d: %v", len(files), files)
	}
	if root != dir {
		t.Errorf("sourceRoot: want %q, got %q", dir, root)
	}
}

func TestResolveUploadFiles_DirectoryTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "a")

	files, root := resolveUploadFiles(dir + "/")
	if len(files) != 1 {
		t.Errorf("want 1 file, got %d", len(files))
	}
	if root != dir {
		t.Errorf("sourceRoot: want %q, got %q", dir, root)
	}
}

func TestResolveUploadFiles_NonexistentSource(t *testing.T) {
	files, _ := resolveUploadFiles("/nonexistent/path/")
	if len(files) != 0 {
		t.Errorf("want 0 files, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// collectArtifactEntries
// ---------------------------------------------------------------------------

func baseConfig(source string) packages.Config {
	return packages.Config{
		Account:  "acc123",
		PkgURL:   "https://pkg.harness.io",
		Registry: "my-registry",
		Name:     "my-package",
		Version:  "1.0.0",
		Source:   source,
	}
}

func TestCollectArtifactEntries_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.jar")
	writeFile(t, path, "hello")

	entries := collectArtifactEntries(baseConfig(path), "GENERIC")
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Name != "app.jar" {
		t.Errorf("name: want app.jar, got %q", e.Name)
	}
	wantURL := "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/1.0.0/app.jar"
	if e.URL != wantURL {
		t.Errorf("url: want %q, got %q", wantURL, e.URL)
	}
	if e.FilePath != "app.jar" {
		t.Errorf("filePath: want app.jar, got %q", e.FilePath)
	}
	if e.Digest != "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("digest: got %q", e.Digest)
	}
	if e.Registry != "my-registry" {
		t.Errorf("registry: want my-registry, got %q", e.Registry)
	}
	if e.PackageName != "my-package" {
		t.Errorf("packageName: want my-package, got %q", e.PackageName)
	}
	if e.Version != "1.0.0" {
		t.Errorf("version: want 1.0.0, got %q", e.Version)
	}
	if e.PackageType != "GENERIC" {
		t.Errorf("packageType: want GENERIC, got %q", e.PackageType)
	}
}

func TestCollectArtifactEntries_DefaultVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.jar")
	writeFile(t, path, "data")

	config := baseConfig(path)
	config.Version = ""
	entries := collectArtifactEntries(config, "GENERIC")
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Version != "v1" {
		t.Errorf("version: want default v1, got %q", entries[0].Version)
	}
	wantURL := "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/v1/app.jar"
	if entries[0].URL != wantURL {
		t.Errorf("url: want %q, got %q", wantURL, entries[0].URL)
	}
}

func TestCollectArtifactEntries_Directory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "file1.txt"), "one")
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "file2.txt"), "two")

	entries := collectArtifactEntries(baseConfig(dir), "GENERIC")
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	// Directory uploads are stored under the source dir's base name.
	base := filepath.Base(dir)
	byName := map[string]fileArtifactEntry{}
	for _, e := range entries {
		byName[e.Name] = e
		if e.Digest == "" {
			t.Errorf("digest should not be empty for %s", e.Name)
		}
		if e.Registry != "my-registry" || e.PackageName != "my-package" || e.Version != "1.0.0" {
			t.Errorf("HAR fields not populated for %s: %+v", e.Name, e)
		}
	}
	if want := base + "/file1.txt"; byName["file1.txt"].FilePath != want {
		t.Errorf("filePath: want %q, got %q", want, byName["file1.txt"].FilePath)
	}
	if want := base + "/sub/file2.txt"; byName["file2.txt"].FilePath != want {
		t.Errorf("filePath: want %q, got %q", want, byName["file2.txt"].FilePath)
	}
	wantURL := "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/1.0.0/" + base + "/sub/file2.txt"
	if byName["file2.txt"].URL != wantURL {
		t.Errorf("url: want %q, got %q", wantURL, byName["file2.txt"].URL)
	}
}

func TestCollectArtifactEntries_FilenameIgnored(t *testing.T) {
	// hc has no --filename flag (a push with PLUGIN_FILENAME set fails at hc
	// before metadata is written), so metadata always reflects the stored
	// object path: the local file's basename.
	dir := t.TempDir()
	path := filepath.Join(dir, "app.jar")
	writeFile(t, path, "data")

	config := baseConfig(path)
	config.Filename = "renamed.jar"
	entries := collectArtifactEntries(config, "GENERIC")
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].FilePath != "app.jar" {
		t.Errorf("filePath: want app.jar, got %q", entries[0].FilePath)
	}
	wantURL := "https://pkg.harness.io/pkg/acc123/my-registry/files/my-package/1.0.0/app.jar"
	if entries[0].URL != wantURL {
		t.Errorf("url: want %q, got %q", wantURL, entries[0].URL)
	}
}

func TestCollectArtifactEntries_SkipsNonGeneric(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pkg.tgz")
	writeFile(t, path, "data")

	for _, pt := range []string{"NPM", "MAVEN", "PYTHON", "NUGET"} {
		if got := collectArtifactEntries(baseConfig(path), pt); got != nil {
			t.Errorf("want nil for package type %s, got %v", pt, got)
		}
	}
}

func TestCollectArtifactEntries_MissingRequiredConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.jar")
	writeFile(t, path, "data")

	cases := []struct {
		name   string
		mutate func(*packages.Config)
	}{
		{"missing registry", func(c *packages.Config) { c.Registry = "" }},
		{"missing name", func(c *packages.Config) { c.Name = "" }},
		{"missing source", func(c *packages.Config) { c.Source = "" }},
		{"missing pkg URL", func(c *packages.Config) { c.PkgURL = "" }},
		{"missing account", func(c *packages.Config) { c.Account = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := baseConfig(path)
			tc.mutate(&config)
			if got := collectArtifactEntries(config, "GENERIC"); got != nil {
				t.Errorf("want nil, got %v", got)
			}
		})
	}
}

func TestCollectArtifactEntries_NonexistentSource(t *testing.T) {
	if got := collectArtifactEntries(baseConfig("/nonexistent/app.jar"), "GENERIC"); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
