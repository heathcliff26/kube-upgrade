#!/bin/bash

set -e

kubeadm init --pod-network-cidr=10.244.0.0/16 --node-name control-0 --apiserver-cert-extra-sans localhost --apiserver-cert-extra-sans 127.0.0.1
mkdir -p /home/core/.kube
cp -f -i /etc/kubernetes/admin.conf /home/core/.kube/config
chown core:core /home/core/.kube/config
cp -f -i /etc/kubernetes/admin.conf /var/mnt/share/kubeconfig
