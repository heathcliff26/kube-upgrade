package daemon

import (
	"context"
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateNodeStatus(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Node  *corev1.Node
		Error bool
	}{
		{
			Name: "Success",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testnode",
					Annotations: map[string]string{
						constants.KubernetesUpgradeStatus: "unset",
					},
				},
			},
		},
		{
			Name: "NoAnnotations",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testnode",
				},
			},
		},
		{
			Name:  "NoNode",
			Error: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			c := fake.NewSimpleClientset()
			d := &daemon{
				client: c,
				ctx:    context.Background(),
			}
			if tCase.Node != nil {
				_, _ = c.CoreV1().Nodes().Create(context.Background(), tCase.Node, metav1.CreateOptions{})
				d.node = tCase.Node.GetName()
			} else {
				d.node = "not-a-node"
			}

			if tCase.Error {
				assert.Error(d.updateNodeStatus("new-status"), "Should fail")
			} else {
				assert.NoError(d.updateNodeStatus("new-status"), "Should succeed")
				node, _ := c.CoreV1().Nodes().Get(context.Background(), d.node, metav1.GetOptions{})
				assert.Equal("new-status", node.GetAnnotations()[constants.KubernetesUpgradeStatus], "Should have set status")
			}
		})
	}
}
