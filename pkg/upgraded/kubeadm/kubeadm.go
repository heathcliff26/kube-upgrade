package kubeadm

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

const (
	tmpDir = "/tmp/upgraded"
)

type KubeadmCMD struct {
	binary  string
	chroot  string
	mutex   sync.Mutex
	version string
}

// Create a new wrapper for kubeadm.
// The binary will run in the provided chroot.
func NewFromPath(chroot, path string) (*KubeadmCMD, error) {
	err := utils.CheckExistsAndIsExecutable(chroot + path)
	if err != nil {
		return nil, err
	}

	k := &KubeadmCMD{
		binary: path,
		chroot: chroot,
	}

	k.version, err = k.getVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeadm version: %v", err)
	}
	k.version, _ = strings.CutSuffix(k.version, "\n")

	return k, nil
}

// Download kubeadm and create a launch wrapper for it.
func NewFromVersion(chroot, version string) (*KubeadmCMD, error) {
	slog.Info("Downloading kubeadm", slog.String("version", version))

	kubeadmPath := tmpDir + "/kubeadm-" + version
	kubeadmPathWithChroot := chroot + kubeadmPath
	baseURL := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/%s/kubeadm", version, runtime.GOARCH)

	err := downloadFile(baseURL, kubeadmPathWithChroot)
	if err != nil {
		return nil, fmt.Errorf("failed to download kubeadm binary: %v", err)
	}
	err = verifyArtifactWithCosign(kubeadmPathWithChroot, baseURL+".sig", baseURL+".cert")
	if err != nil {
		return nil, fmt.Errorf("invalid kubeadm binary: %v", err)
	}
	return NewFromPath(chroot, kubeadmPath)
}

// Run kubeadm upgrade apply
func (k *KubeadmCMD) Apply(version string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	return utils.CreateChrootCMDWithStdout(k.chroot, k.binary, "upgrade", "apply", "--yes", version).Run()
}

// Run kubeadm upgrade node
func (k *KubeadmCMD) Node() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	return utils.CreateChrootCMDWithStdout(k.chroot, k.binary, "upgrade", "node").Run()
}

func (k *KubeadmCMD) Version() string {
	return k.version
}

func (k *KubeadmCMD) getVersion() (string, error) {
	// #nosec G204: Binary path is controlled by the user
	out, err := exec.Command(k.chroot+k.binary, "version", "--output", "short").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
