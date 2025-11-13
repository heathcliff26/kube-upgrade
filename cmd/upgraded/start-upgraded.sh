#!/bin/bash

set -e

# Set the hostname in the pod to the node name.
# This is needed for kubeadm, as it otherwise fails to find the node name.
if [ -z "${NODE_NAME}" ]; then
    echo "Error: NODE_NAME environment variable is empty. It needs to be set to the node name."
    exit 1
fi
hostname "${NODE_NAME}"

# shellcheck disable=SC2068
/upgraded $@
