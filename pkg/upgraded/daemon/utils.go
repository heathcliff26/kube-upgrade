package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Find the node. It will try various methods and confirm them via the machine id.
// The methods are in order:
//   - whoami api call
//   - listing all nodes and iterating over them. (This does not work anymore with k8s 1.32+)
func findNode(client kubernetes.Interface, machineID string) (string, error) {
	node, err := findNodeViaWhoami(client, machineID)
	if err == nil {
		return node, nil
	}
	slog.Info("Failed to find node via whoami call, trying by listing all nodes next", "err", err)

	return findNodeByListingAllNodes(client, machineID)
}

// Call kubernetes auth api and ask whoami
func findNodeViaWhoami(client kubernetes.Interface, machineID string) (string, error) {
	res, err := client.AuthenticationV1().SelfSubjectReviews().Create(context.Background(), &authenticationv1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	name, _ := strings.CutPrefix(res.Status.UserInfo.Username, "system:node:")
	slog.Info("Found username via auth whoami", slog.String("username", res.Status.UserInfo.Username), slog.String("node", name))

	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if node.Status.NodeInfo.MachineID != machineID {
		return "", fmt.Errorf("found node \"%s\", but machine id does not match", name)
	}
	return name, nil
}

// Find the node by iterating over all nodes and comparing machine ids
func findNodeByListingAllNodes(client kubernetes.Interface, machineID string) (string, error) {
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
