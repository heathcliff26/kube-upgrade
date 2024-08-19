#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

pushd "${base_dir}" >/dev/null

# shellcheck source=../vendor/k8s.io/code-generator/kube_codegen.sh
source "vendor/k8s.io/code-generator/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate /dev/null \
    "pkg/apis"

kube::codegen::gen_client \
    --with-watch \
    --with-applyconfig \
    --output-dir "pkg/client" \
    --output-pkg "github.com/heathcliff26/kube-upgrade/pkg/client" \
    --boilerplate /dev/null \
    "pkg/apis"

gobin="${GOBIN:-$(go env GOPATH)/bin}"

echo "Generating manifests"
"${gobin}/controller-gen"   rbac:roleName=upgrade-controller \
                            crd \
                            paths="./..." \
                            output:crd:artifacts:config=manifests/generated \
                            output:rbac:artifacts:config=manifests/generated

popd >/dev/null
