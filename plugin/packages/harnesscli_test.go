// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseHCVersion(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{
			name: "real hc output",
			out:  "hc version 1.3.44\nBuilt with go1.25.0\n",
			want: "1.3.44",
		},
		{
			name: "leading v is stripped",
			out:  "hc version v1.3.44",
			want: "1.3.44",
		},
		{
			name: "prerelease suffix is preserved",
			out:  "hc version 2.0.0-beta.1",
			want: "2.0.0-beta.1",
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
			got, err := parseHCVersion(test.out)
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

func TestHCArchiveName(t *testing.T) {
	tests := []struct {
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{goos: "linux", goarch: "amd64", want: "hc_1.3.44_linux_x86_64.tar.gz"},
		{goos: "linux", goarch: "arm64", want: "hc_1.3.44_linux_arm64.tar.gz"},
		// harness-cli calls macOS "mac-os", not "darwin".
		{goos: "darwin", goarch: "amd64", want: "hc_1.3.44_mac-os_x86_64.tar.gz"},
		{goos: "darwin", goarch: "arm64", want: "hc_1.3.44_mac-os_arm64.tar.gz"},
		{goos: "windows", goarch: "amd64", want: "hc_1.3.44_windows_x86_64.tar.gz"},
		{goos: "windows", goarch: "arm64", want: "hc_1.3.44_windows_arm64.tar.gz"},
		{goos: "windows", goarch: "386", want: "hc_1.3.44_windows_i386.tar.gz"},
		// No 32-bit build is published for these platforms.
		{goos: "linux", goarch: "386", wantErr: true},
		{goos: "darwin", goarch: "386", wantErr: true},
		{goos: "freebsd", goarch: "amd64", wantErr: true},
		{goos: "linux", goarch: "riscv64", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			got, err := hcArchiveName("1.3.44", test.goos, test.goarch)
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

func TestHCCacheBinPathIncludesVersion(t *testing.T) {
	path, err := hcCacheBinPath("1.3.44")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(path, "1.3.44") {
		t.Errorf("cache path %q does not include the version, so a cache hit could be the wrong binary", path)
	}
	if filepath.Base(path) != hcBinaryName() {
		t.Errorf("cache path %q does not end in %q", path, hcBinaryName())
	}
}

func TestResolveFromCandidatesPicksMatchingVersion(t *testing.T) {
	dir := t.TempDir()
	stale := writeFakeHC(t, dir, "stale", "1.3.40")
	match := writeFakeHC(t, dir, "match", "1.3.44")

	got, ok := resolveFromCandidates([]string{stale, match}, "1.3.44")
	if !ok {
		t.Fatal("expected a candidate to be selected")
	}
	if got != match {
		t.Errorf("got %q, want %q: an older hc earlier in the list must not win", got, match)
	}
}

func TestResolveFromCandidatesSkipsMissingAndBrokenEntries(t *testing.T) {
	dir := t.TempDir()
	match := writeFakeHC(t, dir, "match", "1.3.44")

	candidates := []string{
		filepath.Join(dir, "does-not-exist"),
		dir, // a directory, not a binary
		match,
	}

	got, ok := resolveFromCandidates(candidates, "1.3.44")
	if !ok {
		t.Fatal("expected a candidate to be selected")
	}
	if got != match {
		t.Errorf("got %q, want %q", got, match)
	}
}

func TestResolveFromCandidatesRejectsAllMismatches(t *testing.T) {
	dir := t.TempDir()
	stale := writeFakeHC(t, dir, "stale", "1.3.40")

	if got, ok := resolveFromCandidates([]string{stale}, "1.3.44"); ok {
		t.Errorf("expected no match, got %q: a mismatched hc must trigger a download instead", got)
	}
}

func TestResolveHarnessBinHonoursOverride(t *testing.T) {
	dir := t.TempDir()
	override := writeFakeHC(t, dir, "override", "1.3.40")

	t.Setenv(hcPathEnv, override)
	setPinnedForTest(t, "1.3.44")

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
	t.Setenv(hcPathEnv, filepath.Join(t.TempDir(), "missing"))
	setPinnedForTest(t, "1.3.44")

	if _, err := resolveHarnessBin(); err == nil {
		t.Fatal("expected an error when the override does not exist")
	}
}

func TestResolveHarnessBinRequiresPinnedVersion(t *testing.T) {
	t.Setenv(hcPathEnv, "")
	setPinnedForTest(t, "")

	if _, err := resolveHarnessBin(); err == nil {
		t.Fatal("expected an error when no version is pinned")
	}
}

func TestGetHarnessBinNeverReturnsBareName(t *testing.T) {
	dir := t.TempDir()
	override := writeFakeHC(t, dir, "hc", "1.3.44")

	t.Setenv(hcPathEnv, override)
	setPinnedForTest(t, "1.3.44")
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

func TestExtractHCBinary(t *testing.T) {
	const contents = "#!/bin/sh\necho hc\n"

	archive := buildTarGz(t, map[string]string{
		"README.md":     "docs",
		hcBinaryName():  contents,
		"nested/ignore": "other",
	})

	dest := filepath.Join(t.TempDir(), hcBinaryName())
	if err := extractHCBinary(bytes.NewReader(archive), dest); err != nil {
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
}

func TestExtractHCBinaryMissingEntry(t *testing.T) {
	archive := buildTarGz(t, map[string]string{"README.md": "docs"})

	dest := filepath.Join(t.TempDir(), hcBinaryName())
	if err := extractHCBinary(bytes.NewReader(archive), dest); err == nil {
		t.Fatal("expected an error when the archive has no hc entry")
	}
}

// TestEmbeddedHCIsThePinnedVersion is the guard against the embedded archive
// drifting from HC_VERSION: scripts/fetch-hc.sh reads that file, so a bump that
// leaves a stale archive behind must fail here rather than in a pipeline.
func TestEmbeddedHCIsThePinnedVersion(t *testing.T) {
	if len(embeddedHC) == 0 {
		t.Skipf("no harness-cli is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	raw, err := os.ReadFile(filepath.Join("..", "..", "HC_VERSION"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := strings.TrimSpace(string(raw))

	dest := filepath.Join(t.TempDir(), hcBinaryName())
	if err := extractHCBinary(bytes.NewReader(embeddedHC), dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := hcBinaryVersion(dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("embedded harness-cli is %s, but HC_VERSION pins %s: re-run scripts/fetch-hc.sh", got, want)
	}
}

func TestInstallEmbeddedHCRejectsVersionMismatch(t *testing.T) {
	if len(embeddedHC) == 0 {
		t.Skipf("no harness-cli is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	isolateCache(t)

	// The cache path is keyed by the version we claim to want, so extraction
	// succeeds and the version check is what has to catch this.
	if path, err := installEmbeddedHC("0.0.0-not-a-real-release"); err == nil {
		t.Errorf("expected an error, got %q: an embed that disagrees with the pin must not be used", path)
	}
}

// TestResolveHarnessBinWorksOffline covers the case containerless infra now
// relies on: nothing preinstalled, no network, resolution still succeeds.
func TestResolveHarnessBinWorksOffline(t *testing.T) {
	if len(embeddedHC) == 0 {
		t.Skipf("no harness-cli is embedded for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	isolateCache(t)

	raw, err := os.ReadFile(filepath.Join("..", "..", "HC_VERSION"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := strings.TrimSpace(string(raw))

	t.Setenv(hcPathEnv, "")
	t.Setenv("PATH", t.TempDir())
	setPinnedForTest(t, want)

	path, err := resolveHarnessBin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A machine with a matching hc already installed may legitimately win the
	// candidate scan, so assert on the version rather than the location.
	got, err := hcBinaryVersion(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("resolved harness-cli %s at %s, want %s", got, path, want)
	}
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
	previous := pinnedHCVersion
	SetPinnedHCVersion(version)
	t.Cleanup(func() { SetPinnedHCVersion(previous) })
}

// writeFakeHC creates an executable that reports the given version, standing in
// for a real hc during resolution tests.
func writeFakeHC(t *testing.T, dir, name, version string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake executables require a POSIX shell")
	}

	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"hc version " + version + "\"\n"
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
