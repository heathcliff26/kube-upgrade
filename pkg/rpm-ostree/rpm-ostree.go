package rpmostree

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

var binaryPath = "/usr/bin/rpm-ostree"

// Run rpm-ostree and check for new updates
func CheckForUpgrade() (bool, error) {
	cmd := exec.Command(binaryPath, "upgrade", "--check")
	out, err := cmd.CombinedOutput()
	code, ok := getExitCode(err)
	if !ok && err != nil {
		slog.Error("Failed to check for updates with rpm-ostree", "err", err, slog.String("output", string(out)))
		return false, err
	} else if err == nil {
		code = 0
	}

	switch code {
	case 0:
		return true, nil
	case 77:
		return false, nil
	default:
		return false, fmt.Errorf("Exited with unknown exit code %d: %v", code, err)
	}
}

// Upgrade the system using rpm-ostree. Writes command output to stdout/stderr.
//
// WARNING: Will reboot the system when successfull.
func Upgrade() error {
	cmd := exec.Command(binaryPath, "upgrade", "--reboot")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// Rebases the system to the given container image
//
// WARNING: Will reboot the system when successfull.
func Rebase(image string) error {
	cmd := exec.Command(binaryPath, "rebase", "--reboot", "ostree-unverified-registry:"+image)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
