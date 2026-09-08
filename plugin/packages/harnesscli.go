// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"archive/tar"
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
	// hcPathEnv points the plugin at a specific hc binary and skips version
	// resolution entirely. Intended as an escape hatch for air-gapped and
	// self-managed installations that provision hc themselves.
	hcPathEnv = "PLUGIN_HC_PATH"

	// hcReleaseURLTemplate is the harness-cli release archive location.
	hcReleaseURLTemplate = "https://github.com/harness/harness-cli/releases/download/v%s/%s"

	// hcVersionTimeout bounds `hc version` on a candidate binary, so an
	// unrelated executable that happens to be named hc cannot hang the step.
	hcVersionTimeout = 20 * time.Second

	// hcDownloadTimeout bounds fetching the release archive.
	hcDownloadTimeout = 3 * time.Minute
)

// pinnedHCVersion is the harness-cli version this plugin requires, set from the
// embedded HC_VERSION file at startup. Empty means unconfigured, which is a
// build error rather than something to paper over at runtime.
var pinnedHCVersion string

// SetPinnedHCVersion records the required harness-cli version. It is called
// once at startup with the contents of the HC_VERSION file.
func SetPinnedHCVersion(version string) {
	pinnedHCVersion = strings.TrimSpace(version)
}

// PinnedHCVersion returns the harness-cli version this plugin requires.
func PinnedHCVersion() string {
	return pinnedHCVersion
}

var (
	hcOnce sync.Once
	hcPath string
	hcErr  error
)

// getHarnessBin returns an absolute path to an hc binary whose version matches
// the pinned one, downloading it if no suitable binary is already present.
//
// It deliberately never returns a bare "hc": resolving through PATH at exec
// time means an older hc earlier in PATH silently wins, which surfaces as
// confusing "unknown flag" errors rather than a version mismatch.
func getHarnessBin() (string, error) {
	hcOnce.Do(func() {
		hcPath, hcErr = resolveHarnessBin()
	})
	return hcPath, hcErr
}

// resetHarnessBinCache clears the memoized resolution. Used by tests.
func resetHarnessBinCache() {
	hcOnce = sync.Once{}
	hcPath = ""
	hcErr = nil
}

func resolveHarnessBin() (string, error) {
	// An explicit override is honoured as-is; the operator has stated intent.
	// We still report the version so step logs show what actually ran.
	if override := strings.TrimSpace(os.Getenv(hcPathEnv)); override != "" {
		version, err := hcBinaryVersion(override)
		if err != nil {
			return "", fmt.Errorf("%s=%q is not a usable harness-cli binary: %w", hcPathEnv, override, err)
		}
		if version != pinnedHCVersion {
			logrus.Warnf("using harness-cli %s from %s=%s, but this plugin pins %s",
				version, hcPathEnv, override, pinnedHCVersion)
		} else {
			logrus.Debugf("using harness-cli %s from %s=%s", version, hcPathEnv, override)
		}
		return override, nil
	}

	if pinnedHCVersion == "" {
		return "", fmt.Errorf("no pinned harness-cli version is configured; " +
			"the plugin was built without a valid HC_VERSION file")
	}

	if path, ok := resolveFromCandidates(hcCandidatePaths(), pinnedHCVersion); ok {
		logrus.Debugf("using harness-cli %s at %s", pinnedHCVersion, path)
		return path, nil
	}

	// The copy embedded in this binary is the normal path: it needs no network
	// and cannot be a different version than the one we were built against.
	path, embedErr := installEmbeddedHC(pinnedHCVersion)
	if embedErr == nil {
		logrus.Debugf("using embedded harness-cli %s at %s", pinnedHCVersion, path)
		return path, nil
	}
	logrus.Debugf("could not use the embedded harness-cli: %v", embedErr)

	logrus.Infof("harness-cli %s not available locally, downloading", pinnedHCVersion)
	path, err := downloadHC(pinnedHCVersion)
	if err != nil {
		return "", fmt.Errorf("could not obtain harness-cli %s: %w (embedded copy unusable: %v; "+
			"set %s to use a preinstalled binary)", pinnedHCVersion, err, embedErr, hcPathEnv)
	}
	logrus.Debugf("using harness-cli %s at %s", pinnedHCVersion, path)
	return path, nil
}

// installEmbeddedHC unpacks the harness-cli archive embedded in this binary into
// the cache directory and returns the binary's path. The extracted binary is
// version-checked like any other candidate, so a mismatched embed falls through
// to a download rather than running the wrong hc.
func installEmbeddedHC(version string) (string, error) {
	if len(embeddedHC) == 0 {
		return "", fmt.Errorf("this build embeds no harness-cli for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	dest, err := hcCacheBinPath(version)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %w", err)
	}
	if err := extractHCBinary(bytes.NewReader(embeddedHC), dest); err != nil {
		return "", fmt.Errorf("could not extract the embedded harness-cli: %w", err)
	}

	found, err := hcBinaryVersion(dest)
	if err != nil {
		return "", err
	}
	if found != version {
		return "", fmt.Errorf("the embedded harness-cli reports %s, but this plugin pins %s", found, version)
	}
	return dest, nil
}

// resolveFromCandidates returns the first candidate whose reported version
// matches want. Candidates that are missing, unexecutable, or the wrong
// version are skipped with a debug log rather than failing the step.
func resolveFromCandidates(candidates []string, want string) (string, bool) {
	for _, candidate := range candidates {
		version, err := hcBinaryVersion(candidate)
		if err != nil {
			logrus.Debugf("skipping harness-cli candidate %s: %v", candidate, err)
			continue
		}
		if version != want {
			logrus.Debugf("skipping harness-cli candidate %s: version %s, want %s", candidate, version, want)
			continue
		}
		return candidate, true
	}
	return "", false
}

// hcCandidatePaths lists locations to check before unpacking our own copy,
// cheapest first: the conventional install location, then our version-keyed
// cache, then whatever PATH offers. Anything found here still has to report the
// pinned version to be used.
func hcCandidatePaths() []string {
	candidates := []string{bakedHCPath()}

	if cached, err := hcCacheBinPath(pinnedHCVersion); err == nil {
		candidates = append(candidates, cached)
	}

	if found, err := exec.LookPath(hcBinaryName()); err == nil {
		if abs, err := filepath.Abs(found); err == nil {
			candidates = append(candidates, abs)
		}
	}

	return dedupe(candidates)
}

// bakedHCPath is the conventional hc install location. Our own images no longer
// install one there, but self-managed installations and older images do.
func bakedHCPath() string {
	if runtime.GOOS == "windows" {
		return "C:/bin/hc.exe"
	}
	return "/usr/local/bin/hc"
}

func hcBinaryName() string {
	if runtime.GOOS == "windows" {
		return "hc.exe"
	}
	return "hc"
}

// hcCacheBinPath is where we keep downloaded binaries. The version is part of
// the path, so a cache hit cannot be the wrong version.
func hcCacheBinPath(version string) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	if root == "" {
		return "", fmt.Errorf("no cache directory available")
	}
	return filepath.Join(root, "harness", "hc", version, hcBinaryName()), nil
}

var hcVersionPattern = regexp.MustCompile(`version\s+v?(\d+\.\d+\.\d+[^\s]*)`)

// hcBinaryVersion runs `<path> version` and extracts the semantic version.
func hcBinaryVersion(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), hcVersionTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s version failed: %w", path, err)
	}

	version, err := parseHCVersion(string(out))
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return version, nil
}

// parseHCVersion extracts the version from `hc version` output, which looks
// like "hc version 1.3.44\nBuilt with go1.25.0".
func parseHCVersion(out string) (string, error) {
	match := hcVersionPattern.FindStringSubmatch(out)
	if match == nil {
		return "", fmt.Errorf("could not parse harness-cli version from %q", strings.TrimSpace(out))
	}
	return match[1], nil
}

// hcArchiveName maps a Go platform onto a harness-cli release asset name.
// harness-cli names its macOS archives "mac-os" and uses x86_64/i386 rather
// than the Go spellings, so this mapping cannot be a simple interpolation.
func hcArchiveName(version, goos, goarch string) (string, error) {
	var osPart string
	switch goos {
	case "linux":
		osPart = "linux"
	case "darwin":
		osPart = "mac-os"
	case "windows":
		osPart = "windows"
	default:
		return "", fmt.Errorf("harness-cli does not publish builds for %s", goos)
	}

	var archPart string
	switch goarch {
	case "amd64":
		archPart = "x86_64"
	case "arm64":
		archPart = "arm64"
	case "386":
		if goos != "windows" {
			return "", fmt.Errorf("harness-cli does not publish builds for %s/%s", goos, goarch)
		}
		archPart = "i386"
	default:
		return "", fmt.Errorf("harness-cli does not publish builds for %s/%s", goos, goarch)
	}

	return fmt.Sprintf("hc_%s_%s_%s.tar.gz", version, osPart, archPart), nil
}

// downloadHC fetches the pinned release archive and installs the binary into
// the cache directory, returning its path.
func downloadHC(version string) (string, error) {
	archive, err := hcArchiveName(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(hcReleaseURLTemplate, version, archive)

	dest, err := hcCacheBinPath(version)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), hcDownloadTimeout)
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

	if err := extractHCBinary(resp.Body, dest); err != nil {
		return "", fmt.Errorf("could not extract harness-cli from %s: %w", url, err)
	}
	return dest, nil
}

// extractHCBinary pulls the hc executable out of a .tar.gz stream and writes it
// to dest. The write goes to a temporary file first and is then renamed, so a
// concurrent step never observes a partially written binary.
func extractHCBinary(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	wanted := hcBinaryName()
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

func writeExecutable(r io.Reader, dest string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".hc-*")
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
