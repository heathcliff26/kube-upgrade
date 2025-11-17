#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

export REPOSITORY="${REPOSITORY:-ghcr.io/heathcliff26}"
export TAG="${TAG:-latest}"
export KUBE_UPGRADE_NAMESPACE="${KUBE_UPGRADE_NAMESPACE:-kube-upgrade}"

output_dir="${base_dir}/manifests/release"
upgrade_controller_file="${output_dir}/upgrade-controller-helm.yaml"

if [[ "${RELEASE_VERSION}" != "" ]] && [[ "${TAG}" == "latest" ]]; then
    TAG="${RELEASE_VERSION}"
fi

# Need to ensure $schema is not expanded by envsubst
# shellcheck disable=SC2016
export schema='$schema'

# shellcheck disable=SC2155
export kube_upgrade_crd=$(cat "${base_dir}/manifests/generated/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml")
# shellcheck disable=SC2155
export kube_upgrade_rbac_cluster_role=$(envsubst <"${base_dir}/manifests/generated/role.yaml")
# shellcheck disable=SC2155
export kube_upgrade_webhooks=$(envsubst <"${base_dir}/manifests/generated/manifests.yaml")

[ ! -d "${output_dir}" ] && mkdir "${output_dir}"

echo "Creating upgrade-controller deployment"
envsubst <"${base_dir}/manifests/base/upgrade-controller.yaml.template" >"${output_dir}/upgrade-controller.yaml"

echo "Creating manifest from helm chart"
cat > "${upgrade_controller_file}" <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
EOF
echo "  name: ${KUBE_UPGRADE_NAMESPACE}" >> "${upgrade_controller_file}"
cat "${base_dir}/manifests/generated/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml" >> "${upgrade_controller_file}"

helm template "${base_dir}/manifests/helm" \
    --debug \
    --set fullnameOverride=kube-upgrade \
    --set upgradeController.repository="${REPOSITORY}/kube-upgrade-controller" \
    --set upgradeController.tag="${TAG}" \
    --set upgraded.repository="${REPOSITORY}/kube-upgraded" \
    --set upgraded.tag="${TAG}" \
    --name-template kube-upgrade \
    --namespace "${KUBE_UPGRADE_NAMESPACE}" \
    | grep -v '# Source: kube-upgrade/templates' \
    | grep -v 'helm.sh/chart: kube-upgrade' \
    | grep -v 'app.kubernetes.io/managed-by: Helm' \
    | sed "s/v0.0.0/${TAG}/g" >> "${upgrade_controller_file}"

echo "Fetching latest kubernetes version"
# shellcheck disable=SC2155
export kube_version_latest="$(curl -L -s https://dl.k8s.io/release/stable.txt)"

echo "Creating example plan"
envsubst <"${base_dir}/manifests/base/upgrade-cr.yaml.template" >"${output_dir}/upgrade-cr.yaml"

echo "Wrote manifests to ${output_dir}"

if [ "${TAG}" == "latest" ]; then
    echo "Tag is latest, syncing manifests with examples"
    cp "${output_dir}"/*.yaml "${base_dir}/examples/"
    # TODO Remove: This should be removed when switching manifest generation to helm only
    rm "${base_dir}/examples/upgrade-controller-helm.yaml"
fi
