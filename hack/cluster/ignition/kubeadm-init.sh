#!/bin/bash

set -e

kubeadm init --pod-network-cidr=10.244.0.0/16 --control-plane-endpoint localhost:6443 --node-name control-0
mkdir -p /home/core/.kube
cp -f -i /etc/kubernetes/admin.conf /home/core/.kube/config
chown core:core /home/core/.kube/config
cp -f -i /etc/kubernetes/admin.conf /var/mnt/share/kubeconfig
