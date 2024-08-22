package controller

import (
	"os"
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
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
	assert.Equal(t, &s, p, "Should contain the same string")
	assert.NotSame(t, s, p, "Should not be the same")
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
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusProgressing,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: false,
		},
		{
			Name: "DependenciesComplete",
			Deps: []string{"foo", "foobar"},
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusProgressing,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: false,
		},
		{
			Name: "Wait",
			Deps: []string{"foo", "foobar", "bar"},
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusProgressing,
				"foobar": v1alpha1.PlanStatusComplete,
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
		Status map[string]string
		Result string
	}{
		{
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusComplete,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: v1alpha1.PlanStatusComplete,
		},
		{
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusWaiting,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: v1alpha1.PlanStatusWaiting,
		},
		{
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusComplete,
				"bar":    v1alpha1.PlanStatusProgressing,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: v1alpha1.PlanStatusProgressing,
		},
		{
			Status: map[string]string{
				"foo":    v1alpha1.PlanStatusUnknown,
				"bar":    v1alpha1.PlanStatusProgressing,
				"foobar": v1alpha1.PlanStatusComplete,
			},
			Result: v1alpha1.PlanStatusUnknown,
		},
		{
			Status: map[string]string{},
			Result: v1alpha1.PlanStatusUnknown,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Result, func(t *testing.T) {
			assert.Equal(t, tCase.Result, createStatusSummary(tCase.Status))
		})
	}
}
