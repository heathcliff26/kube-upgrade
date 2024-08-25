package fleetlock

import (
	"strings"

	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
	"k8s.io/klog/v2"
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

// When having // in a URL, it somehow converts the request from POST to GET.
// See: https://github.com/golang/go/issues/69063
// In general it could lead to unintended behaviour.
func TrimTrailingSlash(url string) string {
	res, found := strings.CutSuffix(url, "/")
	if found {
		klog.Warning("Removed trailing slash in URL, as this could lead to undefined behaviour")
	}
	return res
}
