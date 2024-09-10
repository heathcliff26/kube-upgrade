#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${script_dir}" >/dev/null

OUT_DIR="${script_dir}/coverprofiles"
APP="kube-upgrade"

if [ ! -d "${OUT_DIR}" ]; then
    mkdir "${OUT_DIR}"
fi

go test -coverprofile="${OUT_DIR}/cover-${APP}.out.tmp" -coverpkg "./pkg/..." "./pkg/..."
cat "${OUT_DIR}/cover-${APP}.out.tmp" | grep -v "zz_generated" | grep -v "github.com/heathcliff26/kube-upgrade/pkg/client" > "${OUT_DIR}/cover-${APP}.out"
rm "${OUT_DIR}/cover-${APP}.out.tmp"
go tool cover -html "${OUT_DIR}/cover-${APP}.out" -o "${OUT_DIR}/index.html"

popd >/dev/null
