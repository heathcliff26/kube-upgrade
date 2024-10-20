package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDoNodeUpgradeWithRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &daemon{
		client:        fake.NewSimpleClientset(),
		node:          "not-a-node",
		ctx:           ctx,
		retryInterval: time.Second,
	}

	done := make(chan struct{}, 1)
	go func() {
		// Should not panic with nil
		d.doNodeUpgradeWithRetry(nil)
	}()
	t.Cleanup(cancel)

	select {
	case <-done:
		t.Fatal("The upgrade should not succeed")
	case <-time.After(time.Second * 3):
	}
}

func TestDoNodeUpgrade(t *testing.T) {
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
					constants.NodeKubernetesVersion: "v1.31.0",
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

			kubeadmCMD, err := kubeadm.New("testdata/fake-kubeadm.sh")
			if !assert.NoError(err, "Failed to create kubeadm command") {
				t.FailNow()
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			d := &daemon{
				ctx:       ctx,
				fleetlock: client,
				rpmostree: rpmOstreeCMD,
				kubeadm:   kubeadmCMD,
				client:    fake.NewSimpleClientset(),
				node:      "testnode",
			}

			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: d.node,
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
					},
				},
			}
			node, _ = d.client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

			err = d.doNodeUpgrade(node)

			node, _ = d.client.CoreV1().Nodes().Get(context.Background(), node.GetName(), metav1.GetOptions{})

			if succeed {
				assert.NoError(err, "Should exit without error")
				assert.Equal(constants.NodeUpgradeStatusRebasing, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
			} else {
				assert.Error(err, "Should exit with error")
				assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
			}
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
						constants.NodeUpgradeStatus: "unset",
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
				assert.Equal("new-status", node.GetAnnotations()[constants.NodeUpgradeStatus], "Should have set status")
			}
		})
	}
}
