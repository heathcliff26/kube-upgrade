#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${script_dir}" >/dev/null

OUT_DIR="${script_dir}/coverprofiles"

if [ ! -d "${OUT_DIR}" ]; then
    mkdir "${OUT_DIR}"
fi

go test -coverprofile="${OUT_DIR}/cover.out.tmp" -coverpkg "./pkg/..." "./pkg/..."
grep -v "zz_generated" "${OUT_DIR}/cover.out.tmp" | grep -v "github.com/heathcliff26/kube-upgrade/pkg/client" > "${OUT_DIR}/cover.out"
go tool cover -html "${OUT_DIR}/cover.out" -o "${OUT_DIR}/index.html"
rm "${OUT_DIR}/cover.out" "${OUT_DIR}/cover.out.tmp"

popd >/dev/null
