#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
output_dir="${1:-${repo_root}/dist}"
revision="${2:-$(git -C "$repo_root" rev-parse HEAD)}"

if [[ ! "$revision" =~ ^[0-9a-fA-F]{40}$ ]]; then
  echo "revision must be a full 40-character Git SHA" >&2
  exit 1
fi

release_version="${revision:0:12}"
build_time="${BUILD_TIME:-$(date -u +%Y%m%d%H%M%S)}"
go_command="${GO:-go}"

mkdir -p "$output_dir"
output_dir="$(cd -- "$output_dir" && pwd)"
package_root="$(mktemp -d)"
trap 'rm -rf "$package_root"' EXIT

release_artifacts=(
  semx-darwin-amd64.tar.gz
  semx-darwin-arm64.tar.gz
  semx-linux-amd64.tar.gz
  semx-linux-arm64.tar.gz
  semx-windows-amd64.zip
  semx-windows-arm64.zip
)

for release_artifact in "${release_artifacts[@]}" SHA256SUMS; do
  rm -f "${output_dir}/${release_artifact}"
done

build_target() {
  local goos="$1"
  local goarch="$2"
  local binary_name="$3"
  local archive_name="$4"
  local archive_format="$5"
  local target_dir="${package_root}/${goos}-${goarch}"
  local ldflags

  mkdir -p "$target_dir"
  ldflags="-X github.com/popodidi/semx/internal/version.version=${release_version} -X github.com/popodidi/semx/internal/version.commit=${release_version} -X github.com/popodidi/semx/internal/version.buildTime=${build_time}"

  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 "$go_command" -C "$repo_root" build \
    -ldflags "$ldflags" \
    -o "${target_dir}/${binary_name}" \
    ./cmd/semx

  case "$archive_format" in
    tar.gz)
      tar -C "$target_dir" -czf "${output_dir}/${archive_name}" "$binary_name"
      ;;
    zip)
      (cd "$target_dir" && python3 -m zipfile -c "${output_dir}/${archive_name}" "$binary_name")
      ;;
    *)
      echo "unsupported archive format: ${archive_format}" >&2
      exit 1
      ;;
  esac
}

build_target darwin amd64 semx semx-darwin-amd64.tar.gz tar.gz
build_target darwin arm64 semx semx-darwin-arm64.tar.gz tar.gz
build_target linux amd64 semx semx-linux-amd64.tar.gz tar.gz
build_target linux arm64 semx semx-linux-arm64.tar.gz tar.gz
build_target windows amd64 semx.exe semx-windows-amd64.zip zip
build_target windows arm64 semx.exe semx-windows-arm64.zip zip

(
  cd "$output_dir"
  LC_ALL=C sha256sum "${release_artifacts[@]}" > SHA256SUMS
)

printf 'Built Semx release archives for %s in %s\n' "$revision" "$output_dir"
