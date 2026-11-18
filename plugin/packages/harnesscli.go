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
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// cliPathEnv points the plugin at a specific harness-har binary and skips
	// version resolution entirely. Intended as an escape hatch for air-gapped
	// and self-managed installations that provision the CLI themselves.
	cliPathEnv = "PLUGIN_HARNESS_HAR_PATH"

	// legacyHCPathEnv pointed at the retired hc binary. It is only read to warn
	// that it no longer has any effect.
	legacyHCPathEnv = "PLUGIN_HC_PATH"

	// cliReleaseURLTemplate is the harness-har release archive location.
	cliReleaseURLTemplate = "https://github.com/harness/cli/releases/download/har/v%s/%s"

	// cliVersionTimeout bounds `harness-har version` on a candidate binary, so
	// an unrelated executable with the same name cannot hang the step.
	cliVersionTimeout = 20 * time.Second

	// cliDownloadTimeout bounds fetching the release archive.
	cliDownloadTimeout = 3 * time.Minute
)

// pinnedCLIVersion is the harness-har version this plugin requires, set from
// the embedded HC_VERSION file at startup. Empty means unconfigured, which is
// a build error rather than something to paper over at runtime.
var pinnedCLIVersion string

// SetPinnedCLIVersion records the required harness-har version. It is called
// once at startup with the contents of the HC_VERSION file.
func SetPinnedCLIVersion(version string) {
	pinnedCLIVersion = strings.TrimSpace(version)
}

// PinnedCLIVersion returns the harness-har version this plugin requires.
func PinnedCLIVersion() string {
	return pinnedCLIVersion
}

var (
	cliOnce sync.Once
	cliPath string
	cliErr  error
)

// getHarnessBin returns an absolute path to a harness-har binary whose version
// matches the pinned one, downloading it if no suitable binary is present.
//
// It deliberately never returns a bare "harness-har": resolving through PATH
// at exec time means an older binary earlier in PATH silently wins, which
// surfaces as confusing "unknown flag" errors rather than a version mismatch.
func getHarnessBin() (string, error) {
	cliOnce.Do(func() {
		cliPath, cliErr = resolveHarnessBin()
	})
	return cliPath, cliErr
}

// resetHarnessBinCache clears the memoized resolution. Used by tests.
func resetHarnessBinCache() {
	cliOnce = sync.Once{}
	cliPath = ""
	cliErr = nil
}

func resolveHarnessBin() (string, error) {
	if os.Getenv(legacyHCPathEnv) != "" && os.Getenv(cliPathEnv) == "" {
		logrus.Warnf("%s is set but no longer used: the plugin now runs harness-har; set %s instead",
			legacyHCPathEnv, cliPathEnv)
	}

	// An explicit override is honoured as-is; the operator has stated intent.
	// We still report the version so step logs show what actually ran.
	if override := strings.TrimSpace(os.Getenv(cliPathEnv)); override != "" {
		version, err := cliBinaryVersion(override)
		if err != nil {
			return "", fmt.Errorf("%s=%q is not a usable harness-har binary: %w", cliPathEnv, override, err)
		}
		if version != pinnedCLIVersion {
			logrus.Warnf("using harness-har %s from %s=%s, but this plugin pins %s",
				version, cliPathEnv, override, pinnedCLIVersion)
		} else {
			logrus.Debugf("using harness-har %s from %s=%s", version, cliPathEnv, override)
		}
		return override, nil
	}

	if pinnedCLIVersion == "" {
		return "", fmt.Errorf("no pinned harness-har version is configured; " +
			"the plugin was built without a valid HC_VERSION file")
	}

	if path, ok := resolveFromCandidates(cliCandidatePaths(), pinnedCLIVersion); ok {
		logrus.Debugf("using harness-har %s at %s", pinnedCLIVersion, path)
		return path, nil
	}

	// The copy embedded in this binary is the normal path: it needs no network
	// and cannot be a different version than the one we were built against.
	path, embedErr := installEmbeddedCLI(pinnedCLIVersion)
	if embedErr == nil {
		logrus.Debugf("using embedded harness-har %s at %s", pinnedCLIVersion, path)
		return path, nil
	}
	logrus.Debugf("could not use the embedded harness-har: %v", embedErr)

	logrus.Infof("harness-har %s not available locally, downloading", pinnedCLIVersion)
	path, err := downloadCLI(pinnedCLIVersion)
	if err != nil {
		return "", fmt.Errorf("could not obtain harness-har %s: %w (embedded copy unusable: %v; "+
			"set %s to use a preinstalled binary)", pinnedCLIVersion, err, embedErr, cliPathEnv)
	}
	logrus.Debugf("using harness-har %s at %s", pinnedCLIVersion, path)
	return path, nil
}

// installEmbeddedCLI unpacks the harness-har archive embedded in this binary
// into the cache directory and returns the binary's path. The extracted binary
// is version-checked like any other candidate, so a mismatched embed falls
// through to a download rather than running the wrong CLI.
func installEmbeddedCLI(version string) (string, error) {
	if len(embeddedCLI) == 0 {
		return "", fmt.Errorf("this build embeds no harness-har for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	dest, err := cliCacheBinPath(version)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %w", err)
	}
	if err := extractCLIBinary(embeddedCLI, dest); err != nil {
		return "", fmt.Errorf("could not extract the embedded harness-har: %w", err)
	}

	found, err := cliBinaryVersion(dest)
	if err != nil {
		return "", err
	}
	if found != version {
		return "", fmt.Errorf("the embedded harness-har reports %s, but this plugin pins %s", found, version)
	}
	return dest, nil
}

// resolveFromCandidates returns the first candidate whose reported version
// matches want. Candidates that are missing, unexecutable, or the wrong
// version are skipped with a debug log rather than failing the step.
func resolveFromCandidates(candidates []string, want string) (string, bool) {
	for _, candidate := range candidates {
		version, err := cliBinaryVersion(candidate)
		if err != nil {
			logrus.Debugf("skipping harness-har candidate %s: %v", candidate, err)
			continue
		}
		if version != want {
			logrus.Debugf("skipping harness-har candidate %s: version %s, want %s", candidate, version, want)
			continue
		}
		return candidate, true
	}
	return "", false
}

// cliCandidatePaths lists locations to check before unpacking our own copy,
// cheapest first: the conventional install location, then our version-keyed
// cache, then whatever PATH offers. Anything found here still has to report
// the pinned version to be used.
func cliCandidatePaths() []string {
	candidates := []string{bakedCLIPath()}

	if cached, err := cliCacheBinPath(pinnedCLIVersion); err == nil {
		candidates = append(candidates, cached)
	}

	if found, err := exec.LookPath(cliBinaryName()); err == nil {
		if abs, err := filepath.Abs(found); err == nil {
			candidates = append(candidates, abs)
		}
	}

	return dedupe(candidates)
}

// bakedCLIPath is the conventional harness-har install location for
// self-managed images that provision the CLI themselves.
func bakedCLIPath() string {
	if runtime.GOOS == "windows" {
		return "C:/bin/harness-har.exe"
	}
	return "/usr/local/bin/harness-har"
}

func cliBinaryName() string {
	if runtime.GOOS == "windows" {
		return "harness-har.exe"
	}
	return "harness-har"
}

// cliCacheBinPath is where we keep downloaded binaries. The version is part of
// the path, so a cache hit cannot be the wrong version.
func cliCacheBinPath(version string) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	if root == "" {
		return "", fmt.Errorf("no cache directory available")
	}
	return filepath.Join(root, "harness", "harness-har", version, cliBinaryName()), nil
}

var cliVersionPattern = regexp.MustCompile(`version\s+v?(\d+\.\d+\.\d+[^\s]*)`)

// cliBinaryVersion runs `<path> version` and extracts the semantic version.
func cliBinaryVersion(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cliVersionTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s version failed: %w", path, err)
	}

	version, err := parseCLIVersion(string(out))
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return version, nil
}

// parseCLIVersion extracts the version from `harness-har version` output,
// which looks like "harness-har version 1.2.0 (202611041200Z)".
func parseCLIVersion(out string) (string, error) {
	match := cliVersionPattern.FindStringSubmatch(out)
	if match == nil {
		return "", fmt.Errorf("could not parse harness-har version from %q", strings.TrimSpace(out))
	}
	return match[1], nil
}

// cliArchiveName maps a Go platform onto a harness-har release asset name.
// Windows builds ship as zip archives, every other platform as tar.gz.
func cliArchiveName(version, goos, goarch string) (string, error) {
	switch goos {
	case "linux", "darwin", "windows":
	default:
		return "", fmt.Errorf("harness-har is not published for %s", goos)
	}
	switch goarch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("harness-har is not published for %s/%s", goos, goarch)
	}

	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("harness-plugin-har_%s_%s_%s.%s", version, goos, goarch, ext), nil
}

// downloadCLI fetches the pinned release archive and installs the binary into
// the cache directory, returning its path.
func downloadCLI(version string) (string, error) {
	archive, err := cliArchiveName(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(cliReleaseURLTemplate, version, archive)

	dest, err := cliCacheBinPath(version)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cliDownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("could not download %s: unexpected status %s", url, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not download %s: %w", url, err)
	}
	if err := extractCLIBinary(data, dest); err != nil {
		return "", fmt.Errorf("could not extract harness-har from %s: %w", url, err)
	}
	return dest, nil
}

// extractCLIBinary pulls the harness-har executable out of a release archive
// (zip on Windows, tar.gz elsewhere) and writes it to dest. The write goes to a
// temporary file first and is then renamed, so a concurrent step never
// observes a partially written binary.
func extractCLIBinary(archive []byte, dest string) error {
	if bytes.HasPrefix(archive, []byte("PK\x03\x04")) {
		return extractFromZip(archive, dest)
	}
	return extractFromTarGz(archive, dest)
}

func extractFromTarGz(archive []byte, dest string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	defer gz.Close()

	wanted := cliBinaryName()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("archive does not contain %s", wanted)
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != wanted {
			continue
		}
		return writeExecutable(tr, dest)
	}
}

func extractFromZip(archive []byte, dest string) error {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return err
	}

	wanted := cliBinaryName()
	for _, file := range zr.File {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != wanted {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeExecutable(rc, dest)
	}
	return fmt.Errorf("archive does not contain %s", wanted)
}

func writeExecutable(r io.Reader, dest string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".harness-har-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
