# Embedded harness-cli archives

This directory holds the harness-cli release archives that `go:embed` folds into
the plugin binary, one per platform we release for:

```
hc_linux_amd64.tar.gz    hc_darwin_amd64.tar.gz    hc_windows_amd64.tar.gz
hc_linux_arm64.tar.gz    hc_darwin_arm64.tar.gz    hc_windows_arm64.tar.gz
```

The archives are **not** checked in — they are ~14 MB each. Fetch them with:

```bash
scripts/fetch-hc.sh                 # every platform
scripts/fetch-hc.sh linux/amd64     # just the one you are building for
```

The version comes from the repo-root `HC_VERSION` file. `make build`, the
Dockerfiles, and the release pipeline all run the fetch for you; a plain
`go build .` on a fresh clone fails with

```
pattern hcbin/hc_linux_arm64.tar.gz: no matching files found
```

which means you need to run the script first.

Each platform's archive is embedded by the correspondingly named
`../hcembed_<goos>_<goarch>.go`, so a given binary carries only its own
platform's copy.
