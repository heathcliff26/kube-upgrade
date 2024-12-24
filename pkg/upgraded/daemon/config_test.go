package daemon

import (
	"context"
	"testing"
	"time"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateConfigFromNode(t *testing.T) {
	assert := assert.New(t)

	client, err := fleetlock.NewClient("https://fleetlock.example.com", "default")
	if !assert.NoError(err, "Should create a fleetlock client") {
		t.FailNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	d := &daemon{
		ctx:       ctx,
		client:    fake.NewSimpleClientset(),
		node:      "testnode",
		fleetlock: client,
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.node,
			Annotations: map[string]string{
				constants.ConfigStream: "registry.example.com/fcos-k8s",
			},
		},
	}
	_, _ = d.client.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	assert.NoError(d.UpdateConfigFromNode(), "Should succeed")
	assert.Equal("registry.example.com/fcos-k8s", d.stream, "Should have updated stream")

	d.node = "not-a-node"
	assert.Error(d.UpdateConfigFromNode(), "Should fail to update from non-existent node")
}

func TestUpdateConfigFromAnnotations(t *testing.T) {
	tMatrix := []struct {
		Name        string
		Annotations map[string]string
		Result      struct {
			Stream         string
			FleetlockURL   string
			FleetlockGroup string
			CheckInterval  time.Duration
			RetryInterval  time.Duration
		}
		Success bool
	}{
		{
			Name: "AllValues",
			Annotations: map[string]string{
				constants.ConfigStream:         "registry.example.com/fcos-k8s",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "compute",
				constants.ConfigCheckInterval:  "5m",
				constants.ConfigRetryInterval:  "2h",
			},
			Result: struct {
				Stream         string
				FleetlockURL   string
				FleetlockGroup string
				CheckInterval  time.Duration
				RetryInterval  time.Duration
			}{
				Stream:         "registry.example.com/fcos-k8s",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "compute",
				CheckInterval:  5 * time.Minute,
				RetryInterval:  2 * time.Hour,
			},
			Success: true,
		},
		{
			Name: "FleetlockURLOnly",
			Annotations: map[string]string{
				constants.ConfigFleetlockURL: "https://fleetlock.example.com",
			},
			Result: struct {
				Stream         string
				FleetlockURL   string
				FleetlockGroup string
				CheckInterval  time.Duration
				RetryInterval  time.Duration
			}{
				FleetlockURL: "https://fleetlock.example.com",
			},
			Success: true,
		},
		{
			Name:        "MissingFleetlockURL",
			Annotations: map[string]string{},
		},
		{
			Name: "EmptyStream",
			Annotations: map[string]string{
				constants.ConfigStream: "",
			},
		},
		{
			Name: "EmptyFleetlockURL",
			Annotations: map[string]string{
				constants.ConfigFleetlockURL: "",
			},
		},
		{
			Name: "EmptyFleetlockGroup",
			Annotations: map[string]string{
				constants.ConfigFleetlockGroup: "",
			},
		},
		{
			Name: "MisformedCheckInterval",
			Annotations: map[string]string{
				constants.ConfigCheckInterval: "not-a-duration",
			},
		},
		{
			Name: "MisformedRetryInterval",
			Annotations: map[string]string{
				constants.ConfigRetryInterval: "not-a-duration",
			},
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			client, err := fleetlock.NewEmptyClient()
			if !assert.NoError(err, "Should create fleetlock client") {
				t.FailNow()
			}
			d := &daemon{
				fleetlock: client,
			}

			if tCase.Success {
				assert.NoError(d.UpdateConfigFromAnnotations(tCase.Annotations), "Should update config from annotations")

				assert.Equal(tCase.Result.Stream, d.stream, "Stream should match")
				assert.Equal(tCase.Result.FleetlockURL, d.fleetlock.GetURL(), "Fleetlock URL should match")
				assert.Equal(tCase.Result.FleetlockGroup, d.fleetlock.GetGroup(), "Fleetlock group should match")
				assert.Equal(tCase.Result.CheckInterval, d.checkInterval, "Check interval should match")
				assert.Equal(tCase.Result.RetryInterval, d.retryInterval, "Check interval should match")
			} else {
				assert.Error(d.UpdateConfigFromAnnotations(tCase.Annotations), "Should fail to update config from annotations")
			}
		})
	}
}
