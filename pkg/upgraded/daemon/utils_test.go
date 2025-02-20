package daemon

import (
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFindNodeByListingAllNodes(t *testing.T) {
	client := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testnode",
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				MachineID: "1234567890",
			},
		},
	}
	_, _ = client.CoreV1().Nodes().Create(t.Context(), node, metav1.CreateOptions{})

	res, err := findNodeByListingAllNodes(client, "1234567890")

	assert := assert.New(t)

	assert.NoError(err, "Should succeed")
	assert.Equal("testnode", res, "Should find correct node")
}

func TestNodeNeedsUpgrade(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Node   *corev1.Node
		Result bool
	}{
		{
			Name:   "NoAnnotations",
			Node:   &corev1.Node{},
			Result: false,
		},
		{
			Name: "UpdateComplete",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
						constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
					},
				},
			},
			Result: false,
		},
		{
			Name: "MissingVersionAnnotation",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeUpgradeStatus: constants.NodeUpgradeStatusPending,
					},
				},
			},
			Result: false,
		},
		{
			Name: "UpdatePending",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
						constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
					},
				},
			},
			Result: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, nodeNeedsUpgrade(tCase.Node), "Should return expected result")
		})
	}
}
