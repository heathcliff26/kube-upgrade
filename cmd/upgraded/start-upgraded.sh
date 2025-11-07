#!/bin/bash

set -e

# Bind mount the container resolv.conf to prevent issues with systemd-resolved when running kubeadm in chroot
mount --bind /etc/resolv.conf /host/etc/resolv.conf

# Set the hostname in the pod to the node name.
# This is needed for kubeadm, as it otherwise fails to find the node name.
if [ -z "${HOSTNAME}" ]; then
    echo "Error: HOSTNAME environment variable is empty. It needs to be set to the node name."
    exit 1
fi
hostname "${HOSTNAME}"

# shellcheck disable=SC2068
/upgraded $@
