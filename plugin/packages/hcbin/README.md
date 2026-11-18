# Embedded harness-har archives

This directory holds the harness-har release archives (from the `har/v*`
releases of `harness/cli`) that `go:embed` folds into the plugin binary, one
per platform we release for:

```
harness-har_linux_amd64.tar.gz    harness-har_darwin_amd64.tar.gz    harness-har_windows_amd64.zip
harness-har_linux_arm64.tar.gz    harness-har_darwin_arm64.tar.gz    harness-har_windows_arm64.zip
```

The archives are **not** checked in. Fetch them with:

```bash
scripts/fetch-hc.sh                 # every platform
scripts/fetch-hc.sh linux/amd64     # just the one you are building for
```

The version comes from the repo-root `HC_VERSION` file. `make build`, the
Dockerfiles, and the release pipeline all run the fetch for you; a plain
`go build .` on a fresh clone fails with

```
pattern hcbin/harness-har_linux_arm64.tar.gz: no matching files found
```

which means you need to run the script first.

Each platform's archive is embedded by the correspondingly named
`../hcembed_<goos>_<goarch>.go`, so a given binary carries only its own
platform's copy.
