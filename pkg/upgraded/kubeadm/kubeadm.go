package kubeadm

import (
	"sync"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

type KubeadmCMD struct {
	binary string
	mutex  sync.Mutex
}

// Create a new wrapper for kubeadm
func New(path string) (*KubeadmCMD, error) {
	err := utils.CheckExistsAndIsExecutable(path)
	if err != nil {
		return nil, err
	}

	return &KubeadmCMD{
		binary: path,
	}, nil
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
