[![CI](https://github.com/heathcliff26/kube-upgrade/actions/workflows/ci.yaml/badge.svg?event=push)](https://github.com/heathcliff26/kube-upgrade/actions/workflows/ci.yaml)
[![Coverage Status](https://coveralls.io/repos/github/heathcliff26/kube-upgrade/badge.svg)](https://coveralls.io/github/heathcliff26/kube-upgrade)
[![Editorconfig Check](https://github.com/heathcliff26/kube-upgrade/actions/workflows/editorconfig-check.yaml/badge.svg?event=push)](https://github.com/heathcliff26/kube-upgrade/actions/workflows/editorconfig-check.yaml)
[![Generate go test cover report](https://github.com/heathcliff26/kube-upgrade/actions/workflows/go-testcover-report.yaml/badge.svg)](https://github.com/heathcliff26/kube-upgrade/actions/workflows/go-testcover-report.yaml)
[![Renovate](https://github.com/heathcliff26/kube-upgrade/actions/workflows/renovate.yaml/badge.svg)](https://github.com/heathcliff26/kube-upgrade/actions/workflows/renovate.yaml)

# kube-upgrade
Kubernetes controller and daemon for managing cluster updates.

## Table of Contents

- [kube-upgrade](#kube-upgrade)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Usage](#usage)
    - [Prerequisite](#prerequisite)
    - [Installation](#installation)
    - [Manual installation of upgraded](#manual-installation-of-upgraded)
  - [Container Images](#container-images)
    - [Image location](#image-location)
    - [Tags](#tags)
  - [Architecture](#architecture)
    - [upgrade-controller](#upgrade-controller)
    - [upgraded](#upgraded)
  - [Possible problems when upgrading](#possible-problems-when-upgrading)
  - [Links](#links)


## Introduction

## Usage

### Prerequisite

- A kubernetes cluster installed with kubeadm
- Nodes using Fedora CoreOS with **upgraded** already installed (See [Links](#Links))

### Installation

To install kube-upgrade, follow these steps:
1. Create a configuration file for **upgraded** on each node under `/etc/kube-upgraded/config.yaml`. An example config can be found [here](examples/upgraded-config.yaml).
2. Deploy **upgrade-controller**
```
kubectl apply -f https://raw.githubusercontent.com/heathcliff26/kube-upgrade/main/examples/upgrade-controller/upgrade-controller.yaml
```
3. Create the upgrade-plan
```
https://raw.githubusercontent.com/heathcliff26/kube-upgrade/main/examples/upgrade-controller/upgrade-cr.yaml
```

### Manual installation of upgraded

To install upgraded manually on you nodes:
1. Download the binary for your architecture from the latest release
2. Install the binary into the path of your choice
3. Create a systemd service file for upgraded in `/etc/systemd/system/upgraded.service`. An example service can be found [here](examples/upgraded.service).
4. Create a configuration file for upgraded under `/etc/kube-upgraded/config.yaml`. An example config can be found [here](examples/upgraded-config.yaml).
5. Enable the service
```
sudo systemctl daemon-reload
sudo systemctl enable --now upgraded.service
```

## Container Images

### Image location

| Container Registry                                                                             | Image                              |
| ---------------------------------------------------------------------------------------------- | ---------------------------------- |
| [Github Container](https://github.com/users/heathcliff26/packages/container/package/kube-upgrade-controller) | `ghcr.io/heathcliff26/kube-upgrade-controller`   |
| [Docker Hub](https://hub.docker.com/repository/docker/heathcliff26/kube-upgrade-controller)                  | `docker.io/heathcliff26/kube-upgrade-controller` |

### Tags

There are different flavors of the image:

| Tag(s)      | Description                                                 |
| ----------- | ----------------------------------------------------------- |
| **latest**  | Last released version of the image                          |
| **rolling** | Rolling update of the image, always build from main branch. |
| **vX.Y.Z**  | Released version of the image                               |

## Architecture

Kube-upgrade consists of 2 components, the **upgrade-controller** and **upgraded**. They work together to ensure automatic kubernetes updates across your cluster.
It does depend on a fleetlock server to ensure nodes are not updated simoultaneously, as well as for draining nodes beforehand.

**Important Notice**: When creating a plan, it is always necessary to ensure that the control-plane nodes are upgraded first.

### upgrade-controller

The controller runs in the cluster coordinates the upgrades across the cluster by reading the `KubeUpgradePlan` and annotating nodes with the correct settings.
It will do this per group, depending on how the order is defined in the plan.

### upgraded

The upgraded daemon runs on each node and upgrades the node in accordance with the annotations provided by upgrade-controller.

Even without kubernetes version upgrades, it will constantly check for new Fedora CoreOS versions in the same stream and update to them.

When it detects an update for kubernetes, it will execute the following:
1. Reserve a slot with the fleetlock server
2. Rebase the node into the new version using rpm-ostree
3. Run `kubeadm upgrade node` or `kubeadm upgrade apply <version>`, depending on if it is the first node.

## Possible problems when upgrading

So far as i tested, upgrading between patches (e.g. 1.30.3 -> 1.30.4) is going fine. However when upgrading between 1.30 and 1.31, the static pods for kubernetes do not start with a version mismatch (1.30 pod, 1.31 kubelet). This causes the preflight checks to fail. The solution in this case was for me to ignore preflight errors anyway and simply upgrade to 1.31. This fixed the problem.

I think the reason is, that while it is not explicitly stated in the docs (See [Links](#Links)) and kind of hinted it could be done the other way around, the expected way for the upgrade is the following:
1. Upgrade kubeadm to the newest version
2. Run `kubeadm upgrade (node|apply)`
3. Upgrade kubelet

What **upgraded** is doing instead is, it upgrades both kubeadm and kubelet at the same time (by rebasing). So in conclusion:

TLDR; Upgrading between patches (e,g, 1.x.y -> 1.x.z) is fine, upgrading between minor versions (e.g. 1.x -> 1.y) should be done with proper planning, lots of tests, caution and quite possible manually.

## Links

- [Fedora CoreOS Image with kubernetes and upgraded](https://github.com/heathcliff26/containers/tree/main/apps/fcos-k8s)
- [FleetLock server](https://github.com/heathcliff26/fleetlock)
- [Fedora CoreOS](https://fedoraproject.org/coreos/)
- [CoreOS rebasing](https://coreos.github.io/rpm-ostree/container/#rebasing-a-client-system)
- [kubeadm upgrade](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/)
