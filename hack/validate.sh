#!/bin/bash

#!/bin/bash

set -e

script_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)"
base_dir="${script_dir}/.."

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
cp "${base_dir}"/manifests/release/*.yaml "${base_dir}/examples/upgrade-controller/"

rc=0
git update-index --refresh && git diff-index --quiet HEAD -- || rc=1
if [ $rc -ne 0 ]; then
    echo "FATAL: Need to run \"make manifests\" and update the examples with the result"
    exit 1
fi
