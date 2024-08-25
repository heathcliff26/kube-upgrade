package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDoNodeUpgrade(t *testing.T) {
	t.Run("MutexAlreadyHeld", func(t *testing.T) {
		d := &daemon{}

		d.upgrade.Lock()
		t.Cleanup(d.upgrade.Unlock)

		assert.NoError(t, d.doNodeUpgrade(nil), "Should simply exit without doing anything")
	})
	t.Run("LockAlreadyReserved", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusLocked)
		t.Cleanup(func() {
			srv.Close()
		})

		d := &daemon{
			fleetlock: client,
		}
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testnode",
				Annotations: map[string]string{
					constants.KubernetesVersionAnnotation: "v1.31.0",
				},
			},
		}

		err := d.doNodeUpgrade(node)

		assert.ErrorContains(err, "failed to aquire lock:")
	})
	for name, succeed := range map[string]bool{
		"FailedOstreeRebase":    false,
		"SucceededOstreeRebase": true,
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client, srv := NewFakeFleetlockServer(t, http.StatusOK)
			t.Cleanup(func() {
				srv.Close()
			})

			path := "testdata/exit-1.sh"
			if succeed {
				path = "testdata/exit-0.sh"
			}
			rpmOstreeCMD, err := rpmostree.New(path)
			if !assert.NoError(err, "Failed to create rpm-ostree command") {
				t.FailNow()
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			d := &daemon{
				ctx:       ctx,
				fleetlock: client,
				rpmostree: rpmOstreeCMD,
				client:    fake.NewSimpleClientset(),
				node:      "testnode",
			}

			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: d.node,
					Annotations: map[string]string{
						constants.KubernetesVersionAnnotation: "v1.31.0",
					},
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						KubeletVersion: "v1.30.4",
					},
				},
			}
			node, _ = d.client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

			err = d.doNodeUpgrade(node)

			if succeed {
				assert.NoError(err, "Should exit without error")
			} else {
				assert.Error(err, "Should exit with error")
			}
			node, _ = d.client.CoreV1().Nodes().Get(context.Background(), node.GetName(), metav1.GetOptions{})
			assert.Equal(constants.NodeUpgradeStatusRebasing, node.Annotations[constants.KubernetesUpgradeStatus], "Should have set correct node status")
		})
	}
}

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
