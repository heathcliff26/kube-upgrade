#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

COVERPROFILE_OUT_TMP="${base_dir}/coverprofile.out.tmp"
COVERPROFILE_OUT="${base_dir}/coverprofile.out"

# Paths/Patterns to exclude from coverage. Either because they are not relevant
# or because they where auto-generated.
excludes=("zz_generated" "github.com/heathcliff26/kube-upgrade/pkg/client")

go test -v -race -coverprofile=coverprofile.out.tmp -coverpkg "./pkg/..." ./cmd/... ./pkg/...

grep -v -E "$(IFS="|"; echo "${excludes[*]}")" "${COVERPROFILE_OUT_TMP}" > "${COVERPROFILE_OUT}"

rm "${COVERPROFILE_OUT_TMP}"
