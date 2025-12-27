package kubeadm

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
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

func verifySigstoreSignatureCLI(blob, sig, cert string) error {
	return utils.CreateCMDWithStdout("cosign", "verify-blob", blob, "--signature", sig, "--certificate", cert, "--certificate-identity", "krel-staging@k8s-releng-prod.iam.gserviceaccount.com", "--certificate-oidc-issuer", "https://accounts.google.com").Run()
}

func verifyArtifactWithSigstore(blob, sigURL, certURL string) error {
	// This bundle uses public good instance with an added signing key
	trustedRoot, err := root.FetchTrustedRoot()
	if err != nil {
		panic(err)
	}
	sev, err := verify.NewVerifier(trustedRoot, verify.WithSignedCertificateTimestamps(1), verify.WithTransparencyLog(1), verify.WithObserverTimestamps(1))
	if err != nil {
		panic(err)
	}

	return nil
}

func certificateFromURL(url string) (*x509.Certificate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		// not a base64
		out = raw
	}

	certs, err := cryptoutils.UnmarshalCertificatesFromPEM(out)
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certs found in pem file '%s'", url)
	}
	return certs[0], nil
}
