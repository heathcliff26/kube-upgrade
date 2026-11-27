#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname)"
# shellcheck source=cluster/common.sh
source "${base_dir}/hack/cluster/common.sh"

echo "Removing previous cluster, just in case"
make cluster-down

echo "Preparing cluster"
make cluster-sync

export KUBECONFIG="${kubeconfig_file}"

echo "Running e2e test"
go test -count=1 -v -timeout 15m ./tests/e2e/...
