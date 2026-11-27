#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
# shellcheck source=common.sh
source "${base_dir}/hack/cluster/common.sh"

export KUBECONFIG="${kubeconfig_file}"

export REPOSITORY=localhost
# shellcheck disable=SC2155
export TAG="ci-$(date '+%Y%m%d%H%M%S')"

echo "Building container images"
make build

echo "Importing images to cluster"
rm -f "${shared_dir}/controller.tar" "${shared_dir}/upgraded.tar"
podman save -o "${shared_dir}/controller.tar" "${REPOSITORY}/kube-upgrade-controller:${TAG}"
podman save -o "${shared_dir}/upgraded.tar" "${REPOSITORY}/kube-upgraded:${TAG}"
"${script_dir}/ssh.sh" sudo podman load -i /var/mnt/share/controller.tar
"${script_dir}/ssh.sh" sudo podman load -i /var/mnt/share/upgraded.tar

echo "Creating manifests"
FLEETLOCK_ENABLED=true make manifests

echo "Deploying controller"
kubectl apply -f "${base_dir}/manifests/release/upgrade-controller.yaml"
kubectl wait -n kube-upgrade deployment kube-upgrade --for condition=Available --timeout 5m
kubectl wait -n kube-upgrade deployment fleetlock --for condition=Available --timeout 5m
