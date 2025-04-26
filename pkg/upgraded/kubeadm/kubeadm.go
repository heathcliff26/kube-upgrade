package kubeadm

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

type KubeadmCMD struct {
	binary  string
	mutex   sync.Mutex
	version string
}

// Create a new wrapper for kubeadm
func New(path string) (*KubeadmCMD, error) {
	err := utils.CheckExistsAndIsExecutable(path)
	if err != nil {
		return nil, err
	}

	k := &KubeadmCMD{
		binary: path,
	}

	k.version, err = k.getVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeadm version: %v", err)
	}
	k.version, _ = strings.CutSuffix(k.version, "\n")

	return k, nil
}

// Run kubeadm upgrade apply
func (k *KubeadmCMD) Apply(version string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	return utils.CreateCMDWithStdout(k.binary, "upgrade", "apply", "--yes", version).Run()
}

// Run kubeadm upgrade node
func (k *KubeadmCMD) Node() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	return utils.CreateCMDWithStdout(k.binary, "upgrade", "node").Run()
}

func (k *KubeadmCMD) Version() string {
	return k.version
}

func (k *KubeadmCMD) getVersion() (string, error) {
	// #nosec G204: Binary path is controlled by the user
	out, err := exec.Command(k.binary, "version", "--output", "short").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
