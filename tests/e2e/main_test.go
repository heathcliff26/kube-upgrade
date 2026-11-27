package e2e

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	err := os.Chdir("../..")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)
	testenv = env.NewWithConfig(cfg)

	exitCode := testenv.Run(m)
	if exitCode != 0 {
		fmt.Printf("Failed e2e testsuite with exit code %d\n", exitCode)
	}
}
