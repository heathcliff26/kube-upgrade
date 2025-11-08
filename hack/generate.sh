#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."
bin_dir="${base_dir}/bin"

export GOBIN="${bin_dir}"

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

echo "Generating manifests"
"${bin_dir}/controller-gen" crd \
    rbac:roleName=upgrade-controller \
    webhook \
    paths="./..." \
    output:crd:artifacts:config=manifests/generated \
    output:rbac:artifacts:config=manifests/generated \
    output:webhook:artifacts:config=manifests/generated

echo "Patching webhook manifests"
yq -i e '.webhooks[0].clientConfig.service.name = "upgrade-controller-webhooks"' manifests/generated/manifests.yaml
# shellcheck disable=SC2016
yq -i e '.webhooks[0].clientConfig.service.namespace = "${KUBE_UPGRADE_NAMESPACE}"' manifests/generated/manifests.yaml
# shellcheck disable=SC2016
yq -i e '.metadata.annotations."cert-manager.io/inject-ca-from" = "${KUBE_UPGRADE_NAMESPACE}/webhook-server-cert"' manifests/generated/manifests.yaml
yq -i e '.metadata.name = "kube-upgrade-webhook"' manifests/generated/manifests.yaml

popd >/dev/null

echo "Generating json schema for yaml validation"
pushd "${base_dir}/manifests/generated" >/dev/null

python3 "${base_dir}/hack/external/openapi2jsonschema.py" kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml

popd >/dev/null
