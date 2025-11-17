#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."
bin_dir="${base_dir}/bin"
helm_dir="${base_dir}/manifests/helm/templates"

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

rm -rf "${base_dir}/manifests/generated"

echo "Generating manifests"
"${bin_dir}/controller-gen" crd \
    rbac:roleName=upgrade-controller \
    webhook \
    paths="./..." \
    output:dir:="manifests/generated"

cp manifests/generated/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml "${helm_dir}/../crds/"

echo "Converting role.yaml to helm template"
echo "{{- if .Values.rbac.create -}}" > "${helm_dir}/role.yaml"
cat manifests/generated/role.yaml \
    | sed 's/namespace: kube-upgrade/namespace: {{ .Release.Namespace }}/g' \
    | sed 's/name: upgrade-controller/name: {{ include "kube-upgrade.fullname" . }}\n  labels:\n    {{- include "kube-upgrade.labels" . | nindent 4 }}/g' \
    >> "${helm_dir}/role.yaml"
echo "{{- end }}" >> "${helm_dir}/role.yaml"

echo "Converting webhooks to helm template"
cp manifests/generated/manifests.yaml "${helm_dir}/webhooks.yaml.tmp"
yq -i e '.metadata.name = "PLACEHOLDER_FULLNAME-webhook"' "${helm_dir}/webhooks.yaml.tmp"
yq -i e '.metadata.annotations."cert-manager.io/inject-ca-from" = "PLACEHOLDER_NAMESPACE/PLACEHOLDER_FULLNAME-webhook-cert"' "${helm_dir}/webhooks.yaml.tmp"
yq -i e '.metadata.labels."PLACEHOLDER_LABELS" = "true"' "${helm_dir}/webhooks.yaml.tmp"
yq -i e '.webhooks[0].clientConfig.service.name = "PLACEHOLDER_FULLNAME-webhooks"' "${helm_dir}/webhooks.yaml.tmp"
yq -i e '.webhooks[0].clientConfig.service.namespace = "PLACEHOLDER_NAMESPACE"' "${helm_dir}/webhooks.yaml.tmp"
echo "{{- if .Values.webhooks.enabled -}}" > "${helm_dir}/webhooks.yaml"
cat "${helm_dir}/webhooks.yaml.tmp" \
    | sed 's/PLACEHOLDER_FULLNAME/{{ include "kube-upgrade.fullname" . }}/g' \
    | sed 's/PLACEHOLDER_NAMESPACE/{{ .Release.Namespace }}/g' \
    | sed 's/PLACEHOLDER_LABELS: "true"/{{- include "kube-upgrade.labels" . | nindent 4 }}/g' \
    >> "${helm_dir}/webhooks.yaml"
echo "{{- end }}" >> "${helm_dir}/webhooks.yaml"
rm "${helm_dir}/webhooks.yaml.tmp"

echo "Patching role to allow for customizing namespace"
# shellcheck disable=SC2016
sed -i 's/namespace: kube-upgrade/namespace: ${KUBE_UPGRADE_NAMESPACE}/g' manifests/generated/role.yaml

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
