#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"

if ! command -v kubeconform >/dev/null 2>&1; then
    go install github.com/yannh/kubeconform/cmd/kubeconform@latest
fi

echo "Check if source code is formatted"
make fmt
rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make fmt\"" >&2
    exit 1
fi

echo "Check if go.mod and vendor are up to date"
make update-deps
rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make update-deps\"" >&2
    exit 1
fi

echo "Check if the auto generated code is up to date"
"${script_dir}/generate.sh"

rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make generate\"" >&2
    exit 1
fi

echo "Check if the example manifests are up to date"
export TAG="latest"
export RELEASE_VERSION=""
"${script_dir}/manifests.sh"

git update-index --refresh || echo "examples/upgrade-controller/upgrade-cr.yaml might only have a kubernetes version update, which will be ignored"
rc=0
git diff-index -I "kubernetesVersion: v1.*" --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make manifests\" and update the examples with the result" >&2
    exit 1
fi

# Ensure an updated kubernetes version is not detected as a problem further down.
git checkout examples/upgrade-controller/upgrade-cr.yaml

echo "Check if the generated example plan is conform"
#kubeconform -schema-location manifests/generated/kubeupgradeplan_v1alpha2.json -verbose -strict examples/upgrade-controller/upgrade-cr.yaml
echo "skipping kubeconform. Check if it is correctly managing semver again"
