package controller

import (
	"os"
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestGetNamespace(t *testing.T) {
	oldPath := serviceAccountNamespaceFile

	t.Run("FallbackName", func(t *testing.T) {
		serviceAccountNamespaceFile = "not-a-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.Nil(err)
		assert.Equal(namespaceKubeUpgrade, ns)
	})
	t.Run("ReadFromFile", func(t *testing.T) {
		serviceAccountNamespaceFile = "ns-from-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		err := os.WriteFile(serviceAccountNamespaceFile, []byte("success"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			err = os.Remove(serviceAccountNamespaceFile)
			t.Log(serviceAccountNamespaceFile)
			if err != nil {
				t.Log(err)
			}
		})

		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.Nil(err)
		assert.Equal("success", ns)
	})
	t.Run("FileEmpty", func(t *testing.T) {
		serviceAccountNamespaceFile = "ns-from-empty-file"
		t.Cleanup(func() {
			serviceAccountNamespaceFile = oldPath
		})
		err := os.WriteFile(serviceAccountNamespaceFile, []byte(""), 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			err = os.Remove(serviceAccountNamespaceFile)
			t.Log(serviceAccountNamespaceFile)
			if err != nil {
				t.Log(err)
			}
		})

		assert := assert.New(t)

		ns, err := GetNamespace()
		assert.Error(err)
		assert.Equal("", ns)
	})
}

func TestPointer(t *testing.T) {
	s := "test"
	p := Pointer(s)
	assert.Equal(t, &s, p, "Should return pointer to variable with the same value")
}

func TestGroupWaitForDependency(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Deps   []string
		Status map[string]string
		Result bool
	}{
		{
			Name: "NoDependencies",
			Deps: nil,
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: false,
		},
		{
			Name: "DependenciesComplete",
			Deps: []string{"foo", "foobar"},
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: false,
		},
		{
			Name: "Wait",
			Deps: []string{"foo", "foobar", "bar"},
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, groupWaitForDependency(tCase.Deps, tCase.Status))
		})
	}
}

func TestCreateStatusSummary(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Status map[string]string
		Result string
	}{
		{
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusComplete,
				"foobar": api.PlanStatusComplete,
			},
			Result: api.PlanStatusComplete,
		},
		{
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusWaiting,
				"foobar": api.PlanStatusComplete,
			},
			Result: api.PlanStatusWaiting,
		},
		{
			Name: api.PlanStatusProgressing,
			Status: map[string]string{
				"foo":    api.PlanStatusComplete,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: api.PlanStatusProgressing + ": Upgrading groups [bar]",
		},
		{
			Status: map[string]string{
				"foo":    api.PlanStatusUnknown,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: api.PlanStatusUnknown,
		},
		{
			Name: api.PlanStatusError,
			Status: map[string]string{
				"foo":    api.PlanStatusError,
				"bar":    api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				"foobar": api.PlanStatusComplete,
			},
			Result: api.PlanStatusError + ": Some groups encountered errors [foo]",
		},
		{
			Name:   "EmptyStatus",
			Status: map[string]string{},
			Result: api.PlanStatusUnknown,
		},
	}

	for _, tCase := range tMatrix {
		if tCase.Name == "" {
			tCase.Name = tCase.Result
		}
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, createStatusSummary(tCase.Status))
		})
	}
}

func TestGetUpgradedImage(t *testing.T) {
	tMatrix := []struct {
		Name, ImageEnv, TagEnv, Expected string
	}{
		{
			Name:     "BothEnvSet",
			ImageEnv: "registry.example.com/upgraded-custom",
			TagEnv:   "v1.2.3",
			Expected: "registry.example.com/upgraded-custom:v1.2.3",
		},
		{
			Name:     "OnlyImageEnvSet",
			ImageEnv: "registry.example.com/upgraded-custom",
			TagEnv:   "",
			Expected: "registry.example.com/upgraded-custom:" + version.Version(),
		},
		{
			Name:     "OnlyTagEnvSet",
			ImageEnv: "",
			TagEnv:   "v1.2.3",
			Expected: defaultUpgradedImage + ":v1.2.3",
		},
		{
			Name:     "NoEnvSet",
			ImageEnv: "",
			TagEnv:   "",
			Expected: defaultUpgradedImage + ":" + version.Version(),
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			if tCase.ImageEnv != "" {
				t.Setenv(upgradedImageEnv, tCase.ImageEnv)
			}
			if tCase.TagEnv != "" {
				t.Setenv(upgradedTagEnv, tCase.TagEnv)
			}

			assert.Equal(t, tCase.Expected, GetUpgradedImage(), "Upgraded image should match expected value")
		})
	}
}
