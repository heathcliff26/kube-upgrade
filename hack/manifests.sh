#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

export REPOSITORY="${REPOSITORY:-ghcr.io/heathcliff26}"
export TAG="${TAG:-latest}"
export KUBE_UPGRADE_NAMESPACE="${KUBE_UPGRADE_NAMESPACE:-kube-upgrade}"

output_dir="${base_dir}/manifests/release"

if [[ "${RELEASE_VERSION}" != "" ]] && [[ "${TAG}" == "latest" ]]; then
    TAG="${RELEASE_VERSION}"
fi

# shellcheck disable=SC2155
export kube_upgrade_crd=$(cat "${base_dir}/manifests/generated/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml")
# shellcheck disable=SC2155
export kube_upgrade_rbac_cluster_role=$(cat "${base_dir}/manifests/generated/role.yaml")

[ ! -d "${output_dir}" ] && mkdir "${output_dir}"

echo "Creating upgrade-controller deployment"
envsubst < "${base_dir}/manifests/base/upgrade-controller.yaml.template" > "${output_dir}/upgrade-controller.yaml"

echo "Fetching latest kubernetes version"
# shellcheck disable=SC2155
export kube_version_latest="$(curl -L -s https://dl.k8s.io/release/stable.txt)"

echo "Creating example plan"
envsubst < "${base_dir}/manifests/base/upgrade-cr.yaml.template" > "${output_dir}/upgrade-cr.yaml"

echo "Wrote manifests to ${output_dir}"

if [ "${TAG}" == "latest" ]; then
    echo "Tag is latest, syncing manifests with examples"
    cp "${output_dir}"/*.yaml "${base_dir}/examples/upgrade-controller/"
fi
