package kubeadm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
)

var CosignBinary = "cosign"

func downloadFile(url, dest string) error {
	dir := filepath.Dir(dest)
	// #nosec G301: The binary is no secret, can be world readable/executable
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %v", dir, err)
	}

	// #nosec G107: Yes, the url is variable, that is indeed intended.
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download '%s', received status code %d", url, res.StatusCode)
	}

	// #nosec G304: The path is variable, that is indeed intended.
	// #nosec G302: The binary is no secret, can be world readable/executable
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
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

func verifyArtifactWithCosign(blob, sig, cert string) error {
	args := []string{
		"verify-blob", blob,
		"--signature", sig,
		"--certificate", cert,
		"--certificate-identity", "krel-staging@k8s-releng-prod.iam.gserviceaccount.com",
		"--certificate-oidc-issuer", "https://accounts.google.com",
	}
	return utils.CreateCMDWithStdout(CosignBinary, args...).Run()
}
