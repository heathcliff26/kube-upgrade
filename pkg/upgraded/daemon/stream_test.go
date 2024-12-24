package daemon

import (
	"context"
	"net/http"
	"testing"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDoUpgrade(t *testing.T) {
	fakeDaemon := func(fleetlock *fleetlock.FleetlockClient, rpmostree *rpmostree.RPMOStreeCMD) *daemon {
		d := &daemon{
			fleetlock: fleetlock,
			rpmostree: rpmostree,
			node:      "testnode",
			client:    fake.NewSimpleClientset(),
		}

		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: d.node,
			},
		}
		_, _ = d.client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

		return d
	}

	t.Run("LockAlreadyReserved", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusLocked)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.ErrorContains(err, "failed to aquire lock:")
	})
	t.Run("FailedOstreeUpgrade", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-1.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.Error(err, "Should exit with error")
	})
	// This case is kinda scetchy, as in reality the system would reboot on success, thus the method should never return
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.NoError(err, "Should succeed")
	})
}
