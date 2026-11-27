#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
# shellcheck source=common.sh
source "${base_dir}/hack/cluster/common.sh"

export KUBECONFIG="${kubeconfig_file}"

export REPOSITORY=localhost
export TAG=ci

echo "Building container images"
make build

echo "Importing images to cluster"
podman save -o "${shared_dir}/controller.tar" localhost/kube-upgrade-controller:ci
podman save -o "${shared_dir}/upgraded.tar" localhost/kube-upgraded:ci
"${script_dir}/ssh.sh" sudo podman load -i /var/mnt/share/controller.tar
"${script_dir}/ssh.sh" sudo podman load -i /var/mnt/share/upgraded.tar

echo "Creating manifests"
make manifests

echo "Deploying controller"
kubectl apply -f "${base_dir}/manifests/release/upgrade-controller.yaml"
kubectl wait -n kube-upgrade deployment kube-upgrade --for condition=Available --timeout 5m
