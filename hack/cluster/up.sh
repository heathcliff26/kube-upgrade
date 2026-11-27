#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
# shellcheck source=common.sh
source "${base_dir}/hack/cluster/common.sh"

vm_image="${vm_dir}/fcos.qcow2"
ignition_file="${vm_dir}/control.ign"
butane_file="${vm_dir}/control.bu"

fcos_url="https://builds.coreos.fedoraproject.org/prod/streams/stable/builds/45.20261030.3.1/x86_64/fedora-coreos-45.20261030.3.1-qemu.x86_64.qcow2.xz"

if [ -f "${pid_file}" ]; then
    if kill -0 "$(cat "${pid_file}")" 2>/dev/null; then
        echo "Cluster is already running, exiting..."
        exit 0
    fi
fi

[ -d "${shared_dir}" ] && rm -rf "${shared_dir}"
mkdir -p "${shared_dir}"

if [ ! -f "${vm_image}" ]; then
    echo "Downloading Fedora CoreOS"
    curl -SL -o "${vm_image}.xz" "${fcos_url}"
    xz -d "${vm_image}.xz"
fi

if [ ! -f "${ssh_key}" ]; then
    echo "Creating SSH Key"
    ssh-keygen -q -t ed25519 -C "kube-upgrade" -f "${ssh_key}" -N ""
fi
ssh_public_key="$(cat "${ssh_key}.pub")"

sed "s#SSH_PUBLIC_KEY#${ssh_public_key}#g" "${script_dir}/ignition/control.bu" > "${butane_file}"

alias butane='podman run --rm --interactive         \
                --security-opt label=disable          \
                --volume "${PWD}":/pwd --workdir /pwd \
                quay.io/coreos/butane:release'

pushd "${script_dir}"
podman run --rm -i \
    --security-opt label=disable \
    --volume "${PWD}":/pwd --workdir /pwd \
    quay.io/coreos/butane:release \
    -p -o "vm/control.ign" -s "vm/control.bu" -d ./ignition
popd

echo "Starting VM"
qemu-kvm -m 4096 -cpu host -smp 4 -snapshot -daemonize -enable-kvm \
    -display none \
    -pidfile "${pid_file}" \
    -chardev file,id=serial0,path="${vm_dir}/console.log" \
    -serial chardev:serial0 \
    -drive "if=virtio,file=${vm_image}" \
    -fw_cfg "name=opt/com.coreos/config,file=${ignition_file}" \
    -virtfs "local,path=${shared_dir},mount_tag=shared,security_model=mapped-xattr" \
    -nic user,model=virtio,hostfwd=tcp::2222-:22,hostfwd=tcp::6443-:6443

echo "Waiting for cluster to be available"
for i in {1..10}; do
    echo "Checking if cluster is up: ${i}/10"
    if [ -f "${kubeconfig_file}" ]; then
        break
    fi
    if ! kill -0 "$(cat "${pid_file}")" 2>/dev/null; then
        echo "VM is no longer running, did it crash?"
        exit 1
    fi
    sleep 30
done

if [ ! -f "${kubeconfig_file}" ]; then
    echo "Timed out waiting for cluster, VM might still be running in the background"
    exit 1
fi

export KUBECONFIG=${kubeconfig_file}

echo "Patching kubeconfig"
kubectl config set clusters.kubernetes.server https://localhost:6443

echo "Making node schedulable"
kubectl taint nodes control-0 node-role.kubernetes.io/control-plane:NoSchedule-

echo "Fixing node label"
kubectl label node control-0 --overwrite node-role.kubernetes.io/control-plane=true

echo "Creating CNI"
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
kubectl rollout status -n kube-flannel daemonset kube-flannel-ds --timeout 5m
# Since we are single node, coredns needs to keep running on the node even during the upgrade
# This will also restart the pods, which is needed after the CNI is deployed.
kubectl patch deployment -n kube-system coredns --type='json' -p='[
    {
        "op": "add",
        "path": "/spec/template/spec/tolerations/-",
        "value": {
        "key": "node.kubernetes.io/unschedulable",
        "operator": "Exists",
        "effect": "NoSchedule"
        }
    }
]'

echo "Fixing cni0 network interface"
# The interface has sometimes the wrong ip, causing problems.
# Remove it, as flannel does not need it.
"${script_dir}/ssh.sh" sudo ip link set cni0 down
"${script_dir}/ssh.sh" sudo ip link delete cni0

echo "Deploying cert-manager"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.21.3/cert-manager.yaml
kubectl wait -n cert-manager deployment cert-manager --for condition=Available --timeout 5m

echo "Cluster is up"
echo "    export KUBECONFIG=${kubeconfig_file}"
