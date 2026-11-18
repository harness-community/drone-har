// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestCLIAuthEnv(t *testing.T) {
	base := Config{
		Account: "acct",
		Org:     "org1",
		Project: "proj1",
		ApiURL:  "https://app.example.com",
		PkgURL:  "https://pkg.example.com",
	}
	tests := []struct {
		name        string
		token       string
		org         string
		pipelineOrg string
		want        map[string]string
		wantErr     string
	}{
		{
			name:  "PAT uses env mode",
			token: "pat.acct.id.secret",
			want: map[string]string{
				"HARNESS_PROFILE":      "env",
				"HARNESS_API_KEY":      "pat.acct.id.secret",
				"HARNESS_ACCOUNT":      "acct",
				"HARNESS_ORG":          "org1",
				"HARNESS_PROJECT":      "proj1",
				"HARNESS_API_URL":      "https://app.example.com",
				"HARNESS_REGISTRY_URL": "https://pkg.example.com",
			},
		},
		{
			name:  "SAT uses env mode",
			token: "sat.acct.id.secret",
			want: map[string]string{
				"HARNESS_PROFILE":      "env",
				"HARNESS_API_KEY":      "sat.acct.id.secret",
				"HARNESS_ACCOUNT":      "acct",
				"HARNESS_ORG":          "org1",
				"HARNESS_PROJECT":      "proj1",
				"HARNESS_API_URL":      "https://app.example.com",
				"HARNESS_REGISTRY_URL": "https://pkg.example.com",
			},
		},
		{
			name:  "build token uses pipeline mode",
			token: "eyJhbGciOiJIUzI1NiJ9.payload.sig",
			want: map[string]string{
				"HARNESS_PROFILE":     "pipeline",
				"HARNESS_HAR_TOKEN":   "eyJhbGciOiJIUzI1NiJ9.payload.sig",
				"HARNESS_HAR_API_URL": "https://app.example.com",
				"HARNESS_HAR_URL":     "https://pkg.example.com",
				"HARNESS_ACCOUNT_ID":  "acct",
				"HARNESS_ORG_ID":      "org1",
				"HARNESS_PROJECT_ID":  "proj1",
			},
		},
		{
			name:        "pipeline org fills a missing plugin org",
			token:       "pat.acct.id.secret",
			org:         "-",
			pipelineOrg: "pipeline-org",
			want: map[string]string{
				"HARNESS_PROFILE":      "env",
				"HARNESS_API_KEY":      "pat.acct.id.secret",
				"HARNESS_ACCOUNT":      "acct",
				"HARNESS_ORG":          "pipeline-org",
				"HARNESS_PROJECT":      "proj1",
				"HARNESS_API_URL":      "https://app.example.com",
				"HARNESS_REGISTRY_URL": "https://pkg.example.com",
			},
		},
		{
			name:        "plugin org wins over the pipeline org",
			token:       "eyJhbGciOiJIUzI1NiJ9.payload.sig",
			pipelineOrg: "pipeline-org",
			want: map[string]string{
				"HARNESS_PROFILE":     "pipeline",
				"HARNESS_HAR_TOKEN":   "eyJhbGciOiJIUzI1NiJ9.payload.sig",
				"HARNESS_HAR_API_URL": "https://app.example.com",
				"HARNESS_HAR_URL":     "https://pkg.example.com",
				"HARNESS_ACCOUNT_ID":  "acct",
				"HARNESS_ORG_ID":      "org1",
				"HARNESS_PROJECT_ID":  "proj1",
			},
		},
		{
			name:    "missing org is reported with the plugin setting",
			token:   "pat.acct.id.secret",
			org:     "-",
			wantErr: "PLUGIN_ORG",
		},
		{
			name:  "scoped token uses env mode",
			token: "st.acct.id.secret",
			want: map[string]string{
				"HARNESS_PROFILE":      "env",
				"HARNESS_API_KEY":      "st.acct.id.secret",
				"HARNESS_ACCOUNT":      "acct",
				"HARNESS_ORG":          "org1",
				"HARNESS_PROJECT":      "proj1",
				"HARNESS_API_URL":      "https://app.example.com",
				"HARNESS_REGISTRY_URL": "https://pkg.example.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HARNESS_ORG_ID", test.pipelineOrg)
			t.Setenv("HARNESS_PROJECT_ID", "")
			t.Setenv("HARNESS_ACCOUNT_ID", "")
			config := base
			config.Token = test.token
			if test.org == "-" {
				config.Org = ""
			}

			got, err := cliAuthEnv(config)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v\nwant %v", got, test.want)
			}
		})
	}
}

func TestCommandEnvReplacesExistingEntries(t *testing.T) {
	t.Setenv("HARNESS_HAR_TOKEN", "injected-by-the-pipeline")
	t.Setenv("DRONE_HAR_TEST_KEEP", "kept")

	env := commandEnv(map[string]string{"HARNESS_HAR_TOKEN": "from-plugin"})

	var tokens []string
	kept := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "HARNESS_HAR_TOKEN=") {
			tokens = append(tokens, kv)
		}
		if kv == "DRONE_HAR_TEST_KEEP=kept" {
			kept = true
		}
	}
	if !reflect.DeepEqual(tokens, []string{"HARNESS_HAR_TOKEN=from-plugin"}) {
		t.Errorf("HARNESS_HAR_TOKEN entries = %v, want only the plugin's value", tokens)
	}
	if !kept {
		t.Error("unrelated variables must be passed through")
	}
}

func TestBuildPushCommand(t *testing.T) {
	dir := t.TempDir()
	cli := writeFakeCLI(t, dir, "harness-har", "1.2.0")
	t.Setenv(cliPathEnv, cli)
	resetHarnessBinCache()
	t.Cleanup(resetHarnessBinCache)

	base := Config{
		Registry: "reg",
		ApiURL:   "https://app.example.com",
		PomFile:  "pom.xml",
		Metadata: "harness_pipeline_id:p1",
	}
	tests := []struct {
		name        string
		packageType PackageType
		config      func(Config) Config
		withName    bool
		want        []string
	}{
		{
			name:        "generic takes name and version from the plugin",
			packageType: Generic,
			withName:    true,
			want: []string{cli, "push", "artifact:generic", "reg/demo", "file.txt",
				"--version", "v1", "--metadata", "harness_pipeline_id:p1"},
		},
		{
			name:        "maven passes the pom",
			packageType: Maven,
			want: []string{cli, "push", "artifact:maven", "reg", "file.txt",
				"--pom-file", "pom.xml", "--metadata", "harness_pipeline_id:p1"},
		},
		{
			name:        "types without metadata support get none",
			packageType: Conda,
			want:        []string{cli, "push", "artifact:conda", "reg", "file.txt"},
		},
		{
			name:        "local API URLs get no metadata",
			packageType: NPM,
			config:      func(c Config) Config { c.ApiURL = "http://localhost:3000"; return c },
			want:        []string{cli, "push", "artifact:npm", "reg", "file.txt"},
		},
		{
			name:        "no API URL means no metadata",
			packageType: Python,
			config:      func(c Config) Config { c.ApiURL = ""; return c },
			want:        []string{cli, "push", "artifact:python", "reg", "file.txt"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := base
			if test.config != nil {
				config = test.config(config)
			}
			got, err := buildPushCommand(test.packageType, config, "v1", "file.txt", "demo", test.withName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got  %v\nwant %v", got, test.want)
			}
		})
	}
}

func TestParseCLIVersion(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{
			name: "release build",
			out:  "harness-har version 1.2.0 (202611041200Z)\n",
			want: "1.2.0",
		},
		{
			name: "leading v is stripped",
			out:  "harness-har version v1.2.0",
			want: "1.2.0",
		},
		{
			name: "dev build suffix is preserved",
			out:  "harness-har version 0.1.0-dev (dev)",
			want: "0.1.0-dev",
		},
		{
			name:    "unrelated binary",
			out:     "Usage: tar [OPTION...]",
			wantErr: true,
		},
		{
			name:    "empty output",
			out:     "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCLIVersion(test.out)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got version %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestCLIArchiveName(t *testing.T) {
	tests := []struct {
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{goos: "linux", goarch: "amd64", want: "harness-plugin-har_1.2.0_linux_amd64.tar.gz"},
		{goos: "linux", goarch: "arm64", want: "harness-plugin-har_1.2.0_linux_arm64.tar.gz"},
		{goos: "darwin", goarch: "amd64", want: "harness-plugin-har_1.2.0_darwin_amd64.tar.gz"},
		{goos: "darwin", goarch: "arm64", want: "harness-plugin-har_1.2.0_darwin_arm64.tar.gz"},
		// Windows builds ship as zip archives.
		{goos: "windows", goarch: "amd64", want: "harness-plugin-har_1.2.0_windows_amd64.zip"},
		{goos: "windows", goarch: "arm64", want: "harness-plugin-har_1.2.0_windows_arm64.zip"},
		{goos: "windows", goarch: "386", wantErr: true},
		{goos: "linux", goarch: "386", wantErr: true},
		{goos: "freebsd", goarch: "amd64", wantErr: true},
		{goos: "linux", goarch: "riscv64", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			got, err := cliArchiveName("1.2.0", test.goos, test.goarch)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestCLICacheBinPathIncludesVersion(t *testing.T) {
	path, err := cliCacheBinPath("1.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(path, "1.2.0") {
		t.Errorf("cache path %q does not include the version, so a cache hit could be the wrong binary", path)
	}
	if filepath.Base(path) != cliBinaryName() {
		t.Errorf("cache path %q does not end in %q", path, cliBinaryName())
	}
}

func TestResolveFromCandidatesPicksMatchingVersion(t *testing.T) {
	dir := t.TempDir()
	stale := writeFakeCLI(t, dir, "stale", "1.1.0")
	match := writeFakeCLI(t, dir, "match", "1.2.0")

	got, ok := resolveFromCandidates([]string{stale, match}, "1.2.0")
	if !ok {
		t.Fatal("expected a candidate to be selected")
	}
	if got != match {
		t.Errorf("got %q, want %q: an older binary earlier in the list must not win", got, match)
	}
}

func TestResolveFromCandidatesSkipsMissingAndBrokenEntries(t *testing.T) {
	dir := t.TempDir()
	match := writeFakeCLI(t, dir, "match", "1.2.0")

	candidates := []string{
		filepath.Join(dir, "does-not-exist"),
		dir, // a directory, not a binary
		match,
	}

	got, ok := resolveFromCandidates(candidates, "1.2.0")
	if !ok {
		t.Fatal("expected a candidate to be selected")
	}
	if got != match {
		t.Errorf("got %q, want %q", got, match)
	}
}

func TestResolveFromCandidatesRejectsAllMismatches(t *testing.T) {
	dir := t.TempDir()
	stale := writeFakeCLI(t, dir, "stale", "1.1.0")

	if got, ok := resolveFromCandidates([]string{stale}, "1.2.0"); ok {
		t.Errorf("expected no match, got %q: a mismatched binary must trigger a download instead", got)
	}
}

func TestResolveHarnessBinHonoursOverride(t *testing.T) {
	dir := t.TempDir()
	override := writeFakeCLI(t, dir, "override", "1.1.0")

	t.Setenv(cliPathEnv, override)
	setPinnedForTest(t, "1.2.0")

	got, err := resolveHarnessBin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A version mismatch is a warning, not an error: the operator asked for
	// this specific binary.
	if got != override {
		t.Errorf("got %q, want %q", got, override)
	}
}

func TestResolveHarnessBinRejectsUnusableOverride(t *testing.T) {
	t.Setenv(cliPathEnv, filepath.Join(t.TempDir(), "missing"))
	setPinnedForTest(t, "1.2.0")

	if _, err := resolveHarnessBin(); err == nil {
		t.Fatal("expected an error when the override does not exist")
	}
}

func TestResolveHarnessBinRequiresPinnedVersion(t *testing.T) {
	t.Setenv(cliPathEnv, "")
	setPinnedForTest(t, "")

	if _, err := resolveHarnessBin(); err == nil {
		t.Fatal("expected an error when no version is pinned")
	}
}

func TestGetHarnessBinNeverReturnsBareName(t *testing.T) {
	dir := t.TempDir()
	override := writeFakeCLI(t, dir, "harness-har", "1.2.0")

	t.Setenv(cliPathEnv, override)
	setPinnedForTest(t, "1.2.0")
	resetHarnessBinCache()
	t.Cleanup(resetHarnessBinCache)

	got, err := getHarnessBin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("got %q, want an absolute path: a bare name would resolve through PATH at exec time", got)
	}
}

func TestPullCreatesMissingDestination(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(cliPathEnv, writeFakeCLI(t, dir, "harness-har", "1.2.0"))
	setPinnedForTest(t, "1.2.0")
	resetHarnessBinCache()
	t.Cleanup(resetHarnessBinCache)

	dest := filepath.Join(dir, "downloads", "nested")
	config := Config{
		Registry: "reg", Name: "demo", Version: "1.0.0", Filename: "a.txt", Destination: dest,
		Token: "pat.acct.id.secret", Account: "acct", Org: "org", Project: "proj", PkgURL: "https://pkg.example",
	}
	if err := NewGenericHandler().Pull(context.Background(), config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info, err := os.Stat(dest); err != nil || !info.IsDir() {
		t.Errorf("destination %q was not created: %v", dest, err)
	}
}

func TestExtractCLIBinary(t *testing.T) {
	const contents = "#!/bin/sh\necho harness-har\n"
	entries := map[string]string{
		"README.md":     "docs",
		cliBinaryName(): contents,
		"nested/ignore": "other",
	}

	for name, archive := range map[string][]byte{
		"tar.gz": buildTarGz(t, entries),
		"zip":    buildZip(t, entries),
	} {
		t.Run(name, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), cliBinaryName())
			if err := extractCLIBinary(archive, dest); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := os.ReadFile(dest)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != contents {
				t.Errorf("got %q, want %q", got, contents)
			}

			if runtime.GOOS != "windows" {
				info, err := os.Stat(dest)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if info.Mode().Perm()&0o111 == 0 {
					t.Errorf("extracted binary %v is not executable", info.Mode().Perm())
				}
			}
		})
	}
}

func TestExtractCLIBinaryMissingEntry(t *testing.T) {
	entries := map[string]string{"README.md": "docs"}
	for name, archive := range map[string][]byte{
		"tar.gz": buildTarGz(t, entries),
		"zip":    buildZip(t, entries),
	} {
		t.Run(name, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), cliBinaryName())
			if err := extractCLIBinary(archive, dest); err == nil {
				t.Fatal("expected an error when the archive has no harness-har entry")
			}
		})
	}
}

// TestEmbeddedCLIIsThePinnedVersion is the guard against the embedded archive
// drifting from HC_VERSION: scripts/fetch-hc.sh reads that file, so a bump that
// leaves a stale archive behind must fail here rather than in a pipeline.
func TestEmbeddedCLIIsThePinnedVersion(t *testing.T) {
	if len(embeddedCLI) == 0 {
		t.Skipf("no harness-har is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	want := readPinnedVersion(t)

	dest := filepath.Join(t.TempDir(), cliBinaryName())
	if err := extractCLIBinary(embeddedCLI, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := cliBinaryVersion(dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("embedded harness-har is %s, but HC_VERSION pins %s: re-run scripts/fetch-hc.sh", got, want)
	}
}

func TestInstallEmbeddedCLIRejectsVersionMismatch(t *testing.T) {
	if len(embeddedCLI) == 0 {
		t.Skipf("no harness-har is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	isolateCache(t)

	// The cache path is keyed by the version we claim to want, so extraction
	// succeeds and the version check is what has to catch this.
	if path, err := installEmbeddedCLI("0.0.0-not-a-real-release"); err == nil {
		t.Errorf("expected an error, got %q: an embed that disagrees with the pin must not be used", path)
	}
}

// TestResolveHarnessBinWorksOffline covers the case containerless infra now
// relies on: nothing preinstalled, no network, resolution still succeeds.
func TestResolveHarnessBinWorksOffline(t *testing.T) {
	if len(embeddedCLI) == 0 {
		t.Skipf("no harness-har is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	isolateCache(t)

	want := readPinnedVersion(t)
	t.Setenv(cliPathEnv, "")
	t.Setenv("PATH", t.TempDir())
	setPinnedForTest(t, want)

	path, err := resolveHarnessBin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A machine with a matching binary already installed may legitimately win
	// the candidate scan, so assert on the version rather than the location.
	got, err := cliBinaryVersion(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("resolved harness-har %s at %s, want %s", got, path, want)
	}
}

func readPinnedVersion(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "HC_VERSION"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return strings.TrimSpace(string(raw))
}

// isolateCache points os.UserCacheDir at a temporary directory so tests neither
// read nor write the developer's real cache.
func isolateCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LocalAppData", dir)
}

func TestDedupePreservesOrder(t *testing.T) {
	got := dedupe([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// setPinnedForTest sets the pinned version and restores it afterwards.
func setPinnedForTest(t *testing.T, version string) {
	t.Helper()
	previous := pinnedCLIVersion
	SetPinnedCLIVersion(version)
	t.Cleanup(func() { SetPinnedCLIVersion(previous) })
}

// writeFakeCLI creates an executable that reports the given version, standing
// in for a real harness-har during resolution tests.
func writeFakeCLI(t *testing.T, dir, name, version string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake executables require a POSIX shell")
	}

	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"harness-har version " + version + " (test)\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return path
}

func buildTarGz(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, contents := range entries {
		header := &tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(contents)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := tw.Write([]byte(contents)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf.Bytes()
}

func buildZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, contents := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := w.Write([]byte(contents)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf.Bytes()
}
