package e2e

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/utils"
	"sigs.k8s.io/e2e-framework/support/kind"
)

const namespace = "kube-upgrade"

const certManagerVersion = "v1.15.3"

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	clusterName := envconf.RandomName("kube-upgrade-e2e", 24)

	err := os.Chdir("..")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = utils.RunCommandWithSeperatedOutput("make REPOSITORY=localhost TAG="+clusterName+" build", os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = utils.RunCommandWithSeperatedOutput("make REPOSITORY=localhost TAG="+clusterName+" manifests", os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	imageArchive := fmt.Sprintf("tmp_image_%s.tar", clusterName)

	err = utils.RunCommandWithSeperatedOutput(fmt.Sprintf("podman save -o %s localhost/kube-upgrade-controller:%s localhost/kube-upgraded:%s", imageArchive, clusterName, clusterName), os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.LoadImageArchiveToCluster(clusterName, imageArchive),
	)
	testenv.Finish(
		envfuncs.ExportClusterLogs(clusterName, "./logs"),
		envfuncs.DestroyCluster(clusterName),
	)

	exitCode := testenv.Run(m)
	if exitCode != 0 {
		fmt.Printf("Failed e2e testsuite with exit code %d\n", exitCode)
	}

	fmt.Print("\nRunning cleanup\n\n")

	fmt.Printf("Removing image archive file %s\n", imageArchive)
	err = os.Remove(imageArchive)
	if err != nil {
		fmt.Printf("Failed to remove image archive %s: %v\n", imageArchive, err)
		os.Exit(1)
	}

	fmt.Printf("Deleting kind cluster %s\n", clusterName)
	err = utils.RunCommandWithSeperatedOutput("kind delete cluster --name "+clusterName, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("")

	os.Exit(exitCode)
}
