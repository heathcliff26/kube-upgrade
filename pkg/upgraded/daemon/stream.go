package daemon

import (
	"fmt"
	"log/slog"
	"time"
)

// Check for os upgrades and perform them if necessary.
// Runs until context is cancelled
func (d *daemon) watchForUpgrade() {
	var needUpgrade bool
	for {
		d.retry(func() bool {
			var err error
			slog.Debug("Checking for upgrades via rpm-ostree")
			needUpgrade, err = d.rpmostree.CheckForUpgrade()
			if err == nil {
				return true
			}
			slog.Error("Failed to check if there is a new upgrade", "err", err)
			return false
		})

		if needUpgrade {
			slog.Info("New upgrade is necessary, trying to start update")
			d.retry(func() bool {
				err := d.doUpgrade()
				if err == nil {
					return true
				}
				slog.Error("Failed to perform rpm-ostree upgrade", "err", err)
				return false
			})
		} else {
			slog.Debug("No upgrades found")
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(d.checkInterval):
		}
	}
}

// Perform rpm-ostree upgrade
func (d *daemon) doUpgrade() error {
	// There should ever only be one upgrade at a time and any upgrade comes with a reboot anyway.
	// So the best option here is to just silently return if the lock is already held.
	if !d.upgrade.TryLock() {
		return nil
	}
	defer d.upgrade.Unlock()

	err := d.UpdateConfigFromNode()
	if err != nil {
		return fmt.Errorf("failed to update daemon config from node annotations: %v", err)
	}

	err = d.fleetlock.Lock()
	if err != nil {
		return fmt.Errorf("failed to aquire lock: %v", err)
	}

	err = d.rpmostree.Upgrade()
	if err != nil {
		return err
	}

	// This should not be reached, as rpmostree.Upgrade() reboots the node on success.
	// I included it here mainly for completness sake.

	d.releaseLock()
	return nil
}
