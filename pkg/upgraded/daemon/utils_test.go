package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		client := fake.NewSimpleClientset()
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testnode",
			},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{
					MachineID: "1234567890",
				},
			},
		}
		t.Setenv(NodeNameEnv, "testnode")
		_, _ = client.CoreV1().Nodes().Create(t.Context(), node, metav1.CreateOptions{})

		res, err := nodeName(client, "1234567890")

		assert.NoError(err, "Should succeed")
		assert.Equal("testnode", res, "Should return correct node name")
	})
	t.Run("EnviromentVariableEmpty", func(t *testing.T) {
		assert := assert.New(t)

		client := fake.NewSimpleClientset()
		t.Setenv(NodeNameEnv, "")

		res, err := nodeName(client, "1234567890")

		assert.Empty(res, "Should not return a name")
		assert.ErrorContains(err, "NODE_NAME environment variable is empty", "Should return correct error")
	})
	t.Run("NodeDoesNotExist", func(t *testing.T) {
		assert := assert.New(t)

		client := fake.NewSimpleClientset()
		t.Setenv(NodeNameEnv, "testnode")

		res, err := nodeName(client, "1234567890")

		assert.Empty(res, "Should not return a name")
		assert.ErrorContains(err, "failed to get node", "Should return correct error")
	})
	t.Run("MachineIDDoesNotMatch", func(t *testing.T) {
		assert := assert.New(t)

		client := fake.NewSimpleClientset()
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testnode",
			},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{
					MachineID: "abcdefghij",
				},
			},
		}
		t.Setenv(NodeNameEnv, "testnode")
		_, _ = client.CoreV1().Nodes().Create(t.Context(), node, metav1.CreateOptions{})

		res, err := nodeName(client, "1234567890")

		assert.Empty(res, "Should not return a name")
		assert.ErrorContains(err, "does not match host machineID", "Should return correct error")
	})
}

func TestNodeNeedsUpgrade(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Node   *corev1.Node
		Result bool
	}{
		{
			Name:   "NoAnnotations",
			Node:   &corev1.Node{},
			Result: false,
		},
		{
			Name: "UpdateComplete",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
						constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
					},
				},
			},
			Result: false,
		},
		{
			Name: "MissingVersionAnnotation",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeUpgradeStatus: constants.NodeUpgradeStatusPending,
					},
				},
			},
			Result: false,
		},
		{
			Name: "UpdatePending",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
						constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
					},
				},
			},
			Result: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, nodeNeedsUpgrade(tCase.Node), "Should return expected result")
		})
	}
}

func TestDeleteDir(t *testing.T) {
	t.Run("NotExists", func(t *testing.T) {
		assert.NoError(t, deleteDir("/not/an/existing/directory"), "Should do nothing if the directory does not exist")
	})
	t.Run("Exists", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)

		path := t.TempDir()
		err := os.WriteFile(filepath.Join(path, "testfile"), []byte("testdata"), 0644)
		require.NoError(err, "Should create test file successfully")

		assert.NoError(deleteDir(path), "Should delete existing directory without error")
		_, err = os.Stat(path)
		assert.True(os.IsNotExist(err), "Directory should be deleted")
	})
}
