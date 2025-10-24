package rpmostree

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

type RPMOStreeCMD struct {
	binary string
	mutex  sync.Mutex
}

// Create a new wrapper for rpm-ostree
func New(path string) (*RPMOStreeCMD, error) {
	err := utils.CheckExistsAndIsExecutable(path)
	if err != nil {
		return nil, err
	}

	return &RPMOStreeCMD{
		binary: path,
	}, nil
}

// Run rpm-ostree and check for new updates
func (r *RPMOStreeCMD) CheckForUpgrade() (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// #nosec G204: Binary path is controlled by the user
	cmd := exec.Command(r.binary, "upgrade", "--check")
	out, err := cmd.CombinedOutput()
	code, ok := getExitCode(err)
	if !ok && err != nil {
		fmt.Println(string(out))
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
		fmt.Println(string(out))
		return false, fmt.Errorf("rpm-ostree exited with unknown exit code %d", code)
	}
}

// Upgrade the system using rpm-ostree. Writes command output to stdout/stderr.
//
// WARNING: Will reboot the system when successfull.
func (r *RPMOStreeCMD) Upgrade() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return utils.CreateCMDWithStdout(r.binary, "upgrade", "--reboot").Run()
}

// Rebases the system to the given container image
//
// WARNING: Will reboot the system when successfull.
func (r *RPMOStreeCMD) Rebase(image string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return utils.CreateCMDWithStdout(r.binary, "rebase", "--reboot", "ostree-unverified-registry:"+image).Run()
}

// Register upgraded as the driver for updates with rpm-ostree.
// This will prevent the user from calling rpm-ostree directly, unless they bypass the check.
func (r *RPMOStreeCMD) RegisterAsDriver() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return utils.CreateCMDWithStdout(r.binary, "deploy", "--register-driver=upgraded").Run()
}
