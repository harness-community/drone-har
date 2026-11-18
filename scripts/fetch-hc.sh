#!/bin/sh
# Downloads the harness-har release archives that get embedded into the plugin
# binary by plugin/packages/hcembed_*.go.
#
# The version comes from HC_VERSION, the single source of truth. Archives are
# cached; a version bump discards them all rather than leaving a stale mix.
#
# Usage:
#   scripts/fetch-hc.sh                    # every platform we release for
#   scripts/fetch-hc.sh linux/amd64        # just one, for a local build

set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(tr -d '[:space:]' <"$root/HC_VERSION")
dest="$root/plugin/packages/hcbin"
stamp="$dest/.version"

if [ -z "$version" ]; then
	echo "fetch-hc: HC_VERSION is empty" >&2
	exit 1
fi

if [ $# -eq 0 ]; then
	set -- linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
fi

mkdir -p "$dest"

# A version bump invalidates everything already downloaded.
if [ ! -f "$stamp" ] || [ "$(cat "$stamp")" != "$version" ]; then
	rm -f "$dest"/*.tar.gz "$dest"/*.zip
	echo "$version" >"$stamp"
fi

# fetch writes url to the given path, preferring curl and falling back to wget
# so this works in both golang:alpine (busybox wget) and CI images.
fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$2" "$1"
	else
		echo "fetch-hc: neither curl nor wget is available" >&2
		return 1
	fi
}

for platform in "$@"; do
	goos=${platform%/*}
	goarch=${platform#*/}

	case "$goos/$goarch" in
	linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64) ext=tar.gz ;;
	windows/amd64 | windows/arm64) ext=zip ;;
	*)
		echo "fetch-hc: harness-har is not published for $goos/$goarch" >&2
		exit 1
		;;
	esac

	out="$dest/harness-har_${goos}_${goarch}.${ext}"
	if [ -f "$out" ]; then
		continue
	fi

	url="https://github.com/harness/cli/releases/download/har/v${version}/harness-plugin-har_${version}_${goos}_${goarch}.${ext}"
	echo "fetch-hc: $goos/$goarch <- $url"

	# Download to a temporary file so an interrupted run does not leave a
	# truncated archive that a later run would treat as cached.
	tmp="$out.tmp"
	if ! fetch "$url" "$tmp"; then
		rm -f "$tmp"
		echo "fetch-hc: could not download $url" >&2
		exit 1
	fi
	mv "$tmp" "$out"
done
