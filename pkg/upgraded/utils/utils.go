package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Read the machine-id from /etc/machine-id
func GetMachineID() (string, error) {
	b, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	machineID := strings.TrimRight(string(b), "\r\n")
	return machineID, nil
}

// Create a command that writes to stdout/stderr
func CreateCMDWithStdout(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

// Check if the given file exists and is executable
func CheckExistsAndIsExecutable(path string) error {
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	if f.Mode().Perm()&0100 == 0 {
		return fmt.Errorf("%s is not an executable", path)
	}
	return nil
}
