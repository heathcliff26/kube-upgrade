package daemon

import (
	"net/http"
	"testing"

	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
)

func TestDoUpgrade(t *testing.T) {
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

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

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

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

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

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

		err = d.doUpgrade()

		assert.NoError(err, "Should succeed")
	})
}
