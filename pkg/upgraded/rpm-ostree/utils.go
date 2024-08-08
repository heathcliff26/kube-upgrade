package rpmostree

import (
	"os/exec"
)

// Try to extract the exit code from the error.
func getExitCode(err error) (int, bool) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		return exiterr.ExitCode(), true
	}
	return 0, false
}
