#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."
bin_dir="${base_dir}/bin"
helm_dir="${base_dir}/manifests/helm/templates"
generated_dir="${base_dir}/manifests/generated"

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

rm -rf "${generated_dir}"

echo "Generating manifests"
"${bin_dir}/controller-gen" crd \
    rbac:roleName=upgrade-controller \
    webhook \
    paths="./..." \
    output:dir:="${generated_dir}"

cp "${generated_dir}/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml" "${helm_dir}/../crds/"

echo "Converting role.yaml to helm template"
echo "{{- if .Values.rbac.create -}}" > "${helm_dir}/role.yaml"
cat "${generated_dir}/role.yaml" \
    | sed 's/namespace: kube-upgrade/namespace: {{ .Release.Namespace }}/g' \
    | sed 's/name: upgrade-controller/name: {{ include "kube-upgrade.fullname" . }}\n  labels:\n    {{- include "kube-upgrade.labels" . | nindent 4 }}/g' \
    >> "${helm_dir}/role.yaml"
echo "{{- end }}" >> "${helm_dir}/role.yaml"

echo "Patching webhooks with placeholders"
mv "${generated_dir}/manifests.yaml" "${generated_dir}/webhooks.yaml"
yq -i e '.metadata.name = "PLACEHOLDER_FULLNAME-webhook"' "${generated_dir}/webhooks.yaml"
yq -i e '.metadata.annotations."cert-manager.io/inject-ca-from" = "PLACEHOLDER_NAMESPACE/PLACEHOLDER_FULLNAME-webhook-cert"' "${generated_dir}/webhooks.yaml"
yq -i e '.metadata.labels."PLACEHOLDER_LABELS" = "true"' "${generated_dir}/webhooks.yaml"
yq -i e '.webhooks[0].clientConfig.service.name = "PLACEHOLDER_FULLNAME-webhooks"' "${generated_dir}/webhooks.yaml"
yq -i e '.webhooks[0].clientConfig.service.namespace = "PLACEHOLDER_NAMESPACE"' "${generated_dir}/webhooks.yaml"

echo "Converting webhooks to helm template"
echo "{{- if .Values.webhooks.enabled -}}" > "${helm_dir}/webhooks.yaml"
cat "${generated_dir}/webhooks.yaml" \
    | sed 's/PLACEHOLDER_FULLNAME/{{ include "kube-upgrade.fullname" . }}/g' \
    | sed 's/PLACEHOLDER_NAMESPACE/{{ .Release.Namespace }}/g' \
    | sed 's/PLACEHOLDER_LABELS: "true"/{{- include "kube-upgrade.labels" . | nindent 4 }}/g' \
    >> "${helm_dir}/webhooks.yaml"
echo "{{- end }}" >> "${helm_dir}/webhooks.yaml"

popd >/dev/null

echo "Generating json schema for yaml validation"
pushd "${generated_dir}" >/dev/null

python3 "${base_dir}/hack/external/openapi2jsonschema.py" kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml

popd >/dev/null
