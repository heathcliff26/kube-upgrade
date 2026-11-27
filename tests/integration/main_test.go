package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	upgradecontroller "github.com/heathcliff26/kube-upgrade/pkg/upgrade-controller/controller"
	corev1 "k8s.io/api/core/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var repoRoot string
var k8sClient client.Client

func TestMain(m *testing.M) {
	_, currentFile, _, ok := goruntime.Caller(0)
	if !ok {
		fmt.Println("Failed to read repo directory")
		os.Exit(1)
	}
	repoRoot = filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))

	var err error
	var binaryAssetsDir string
	var downloadBinaryAssets bool
	if kubebuilderAssets := os.Getenv("KUBEBUILDER_ASSETS"); kubebuilderAssets != "" {
		binaryAssetsDir = kubebuilderAssets
		downloadBinaryAssets = false
	} else {
		binaryAssetsDir, err = envtest.SetupEnvtestDefaultBinaryAssetsDirectory()
		requireNoError(err, "Should determine envtest binary cache directory")
		downloadBinaryAssets = true
	}

	data, err := os.ReadFile(filepath.Join(repoRoot, "manifests/generated/webhooks.yaml"))
	requireNoError(err, "Should read webhook config")
	webhookConfigData := strings.NewReplacer(
		"PLACEHOLDER_FULLNAME", "kube-upgrade",
		"PLACEHOLDER_NAMESPACE", "default",
	).Replace(string(data))

	f, err := os.CreateTemp(os.TempDir(), "webhooks-*.yaml")
	requireNoError(err, "Should create temp file")
	defer f.Close()
	webhookConfig := f.Name()
	_, err = f.WriteString(webhookConfigData)
	requireNoError(err, "Should write webhook config")
	defer func() {
		err := os.Remove(webhookConfig)
		requireNoError(err, "Should remove webhook config file")
	}()

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join(repoRoot, "manifests/generated/kubeupgrade.heathcliff.eu_kubeupgradeplans.yaml"),
		},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths:            []string{webhookConfig},
			LocalServingHost: "127.0.0.1",
			MaxTime:          10 * time.Second,
			PollInterval:     100 * time.Millisecond,
		},
		BinaryAssetsDirectory:       binaryAssetsDir,
		DownloadBinaryAssets:        downloadBinaryAssets,
		DownloadBinaryAssetsVersion: "1.37.0",
	}

	config, err := testEnv.Start()
	requireNoError(err, "Should start envtest")
	defer func() {
		err := testEnv.Stop()
		requireNoError(err, "Should stop envtest")
	}()

	scheme := k8sruntime.NewScheme()
	requireNoError(api.AddToScheme(scheme), "Should register api scheme")
	requireNoError(clientgoscheme.AddToScheme(scheme), "Should register scheme for client")

	webhookServer := webhook.NewServer(webhook.Options{
		Host:    testEnv.WebhookInstallOptions.LocalServingHost,
		Port:    testEnv.WebhookInstallOptions.LocalServingPort,
		CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
	})
	_ = os.Setenv(upgradecontroller.UpgradedImageEnv, "localhost/kube-upgraded")
	_ = os.Setenv(upgradecontroller.UpgradedTagEnv, "envtest")
	mgr, err := ctrl.NewManager(config, manager.Options{
		Scheme:                 scheme,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: "0",
		Metrics:                server.Options{BindAddress: "0"},
	})
	requireNoError(err, "Should create controller manager")
	requireNoError(upgradecontroller.SetupWithManager(mgr, namespace), "Should setup controller manager")

	managerCtx, cancelManager := context.WithCancel(context.Background())
	managerErr := make(chan error, 1)
	go func() {
		managerErr <- mgr.Start(managerCtx)
	}()
	defer func() {
		cancelManager()
		select {
		case err := <-managerErr:
			requireNoError(err, "Controller manager should stop cleanly")
		case <-time.After(5 * time.Second):
			fmt.Println("Error: controller manager did not stop in time")
		}
	}()

	webhookStarted := false
	for range 100 {
		webhookStarted = webhookServer.StartedChecker()(&http.Request{}) == nil
		if webhookStarted {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !webhookStarted {
		fmt.Println("Fatal: Webhook server did not start in time")
		os.Exit(1)
	}

	k8sClient, err = client.New(config, client.Options{Scheme: scheme})
	requireNoError(err, "Should create API client")

	nodes := []corev1.Node{
		{
			Name: nodeControlName,
			Labels: map[string]string{
				labelControl: "true",
			},
		},
		{
			Name: nodeComputeName,
			Labels: map[string]string{
				labelCompute: "true",
			},
		},
	}
	for _, node := range nodes {
		err := k8sClient.Create(context.Background(), &node)
		requireNoError(err, fmt.Sprintf("Should create node %s", node.GetName()))
	}

	ns := &corev1.Namespace{
		Name: namespace,
	}
	err = k8sClient.Create(context.Background(), ns)
	requireNoError(err, "Should create namespace")

	exitCode := m.Run()
	if exitCode != 0 {
		fmt.Printf("Failed integration testsuite with exit code %d\n", exitCode)
	}
}

func requireNoError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("Fatal: Required no error but received '%v'. Context: %s\n", err, msg)
	os.Exit(1)
}
