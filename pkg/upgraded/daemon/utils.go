package daemon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Find the node by it's machine ID
func findNodeByMachineID(client kubernetes.Interface, machineID string) (string, error) {
	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, node := range nodes.Items {
		if node.Status.NodeInfo.MachineID == machineID {
			return node.GetName(), nil
		}
	}

	return "", fmt.Errorf("found no node with matching machineID: %s", machineID)
}

// Check if the node needs to upgrade it's kubernetes version
func nodeNeedsUpgrade(node *corev1.Node) bool {
	if node.Annotations == nil {
		return false
	}
	status := node.Annotations[constants.NodeUpgradeStatus]
	if status == constants.NodeUpgradeStatusCompleted {
		return false
	}
	if _, ok := node.Annotations[constants.NodeKubernetesVersion]; !ok {
		slog.Warn("Missing version annotation on node", slog.String("node", node.GetName()), slog.String("annotation", constants.NodeKubernetesVersion))
		return false
	}
	return true
}
