# kube-upgrade Helm Chart

This Helm chart deploys the kube-upgrade operator - A Kubernetes Operator for managing cluster updates

## Prerequisites

- Kubernetes 1.32+
- Helm 3.19+
- cert-manager v1.16+ (When using webhooks)
- FluxCD installed in the cluster (recommended)

## Installation

### Installing from OCI Registry (GitHub Packages)

```bash
# Install the chart
helm install kube-upgrade oci://ghcr.io/heathcliff26/manifests/kube-upgrade --version <version>
```

## Configuration

### Minimal Configuration

Just deploy with the default values.

## Values Reference

See [values.yaml](./values.yaml) for all available configuration options.

### Key Parameters

| Parameter                        | Description                         | Default                                        |
| -------------------------------- | ----------------------------------- | ---------------------------------------------- |
| `upgradeController.repository`   | upgrade-controller image repository | `ghcr.io/heathcliff26/kube-upgrade-controller` |
| `upgradeController.tag`          | upgrade-controller image tag        | Same as chart version                          |
| `upgraded.repository`            | upgraded daemon image repository    | `ghcr.io/heathcliff26/kube-upgraded`           |
| `upgraded.tag`                   | upgraded daemon image tag           | Same as chart version                          |
| `upgradeController.replicaCount` | Number of replicas                  | `2`                                            |
| `rbac.create`                    | Create RBAC resources               | `true`                                         |
| `webhooks.enabled`               | Enable webhooks                     | `true`                                         |

## Support

For more information, visit: https://github.com/heathcliff26/kube-upgrade
