#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"

if ! command -v kubeconform 2>&1 >/dev/null; then
    go install github.com/yannh/kubeconform/cmd/kubeconform@latest
fi

echo "Check if source code is formatted"
make fmt
rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make fmt\""
    exit 1
fi

echo "Check if the auto generated code is up to date"
"${script_dir}/generate.sh"

rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make generate\""
    exit 1
fi

echo "Check if the example manifests are up to date"
export TAG="latest"
export RELEASE_VERSION=""
"${script_dir}/manifests.sh"

rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make manifests\" and update the examples with the result"
    exit 1
fi

echo "Check if the generated example plan is conform"
kubeconform -schema-location manifests/generated/kubeupgradeplan_v1alpha2.json -verbose -strict examples/upgrade-controller/upgrade-cr.yaml
