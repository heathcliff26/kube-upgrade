package fleetlock

import (
	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

// Find the machine-id of the current node and generate a zincati appID from it.
func GetZincateAppID() (string, error) {
	machineID, err := utils.GetMachineID()
	if err != nil {
		return "", err
	}

	appID, err := systemdutils.ZincatiMachineID(machineID)
	if err != nil {
		return "", err
	}
	return appID, nil
}
