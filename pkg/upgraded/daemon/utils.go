package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const NodeNameEnv = "NODE_NAME"

// Return the node name by reading NODE_NAME environment variable.
// Return an error if it doesn't match the given machineID.
func nodeName(client kubernetes.Interface, machineID string) (string, error) {
	name := os.Getenv(NodeNameEnv)
	if name == "" {
		return "", fmt.Errorf("NODE_NAME environment variable is empty")
	}

	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get node %s: %v", name, err)
	}

	if node.Status.NodeInfo.MachineID != machineID {
		return "", fmt.Errorf("node '%s' machineID '%s' does not match host machineID '%s'", name, node.Status.NodeInfo.MachineID, machineID)
	}

	return name, nil
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

// Delete the specified directory if it exists
func deleteDir(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return os.RemoveAll(path)
	}
	return nil
}
