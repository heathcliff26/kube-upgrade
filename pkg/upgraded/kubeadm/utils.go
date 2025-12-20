package kubeadm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

func downloadFile(url, dest string) error {
	dir := filepath.Dir(dest)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %v", dir, err)
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %v", dest, err)
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return fmt.Errorf("failed to download '%s' from '%s': %v", dest, url, err)
	}
	return nil
}

func verifySigstoreSignature(blob, sig, cert string) error {
	return utils.CreateCMDWithStdout("cosign", "verify-blob", blob, "--signature", sig, "--certificate", cert, "--certificate-identity", "krel-staging@k8s-releng-prod.iam.gserviceaccount.com", "--certificate-oidc-issuer", "https://accounts.google.com").Run()
}
