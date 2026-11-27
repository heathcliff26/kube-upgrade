#!/bin/bash
# shellcheck disable=SC2034

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath | xargs dirname | xargs dirname)"
script_dir="${base_dir}/hack/cluster"
vm_dir="${script_dir}/vm"
shared_dir="${vm_dir}/share"

ssh_key="${vm_dir}/id_ed25519"
pid_file="${vm_dir}/fcos.pid"
kubeconfig_file="${shared_dir}/kubeconfig"
