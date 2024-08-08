package drain

import (
	"context"
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/utils"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testNodeName  = "Node1"
	testNamespace = "kube-upgrade"
	testPodName   = "Pod1"
)

func initTestCluster(client *fake.Clientset) {
	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName,
		},
	}
	_, _ = client.CoreV1().Nodes().Create(context.Background(), testNode, metav1.CreateOptions{})

	testNS := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(context.Background(), testNS, metav1.CreateOptions{})

	testPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNamespace,
		},
		Spec: v1.PodSpec{
			NodeName:                      testNodeName,
			TerminationGracePeriodSeconds: utils.Pointer(int64(1)),
		},
	}
	_, _ = client.CoreV1().Pods(testNamespace).Create(context.Background(), testPod, metav1.CreateOptions{})
}

func TestNewDrainer(t *testing.T) {
	assert := assert.New(t)

	d, err := NewDrainer(nil)

	assert.Error(err, "Should throw an error when no client is provided")
	assert.Nil(d, "Should not return a drainer no client is provided")

	d, err = NewDrainer(fake.NewSimpleClientset())

	assert.NoError(err, "Should succed")
	assert.NotNil(d, "Should return a drainer")
}

func TestDrainNode(t *testing.T) {
	client := fake.NewSimpleClientset()
	initTestCluster(client)
	d, err := NewDrainer(client)

	assert := assert.New(t)

	if !assert.NoError(err, "Should create a drainer") {
		t.FailNow()
	}

	err = d.DrainNode(testNodeName)

	if !assert.NoError(err, "Should not return an error") {
		t.FailNow()
	}

	node, _ := client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
	assert.True(node.Spec.Unschedulable, "Node should be unschedulable")
}

func TestUncordonNode(t *testing.T) {
	client := fake.NewSimpleClientset()
	initTestCluster(client)
	d, err := NewDrainer(client)

	assert := assert.New(t)

	if !assert.NoError(err, "Should create a drainer") {
		t.FailNow()
	}

	node, _ := client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
	node.Spec.Unschedulable = true
	_, _ = client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})

	err = d.UncordonNode(testNodeName)

	assert.NoError(err, "Should not return an error")

	node, _ = client.CoreV1().Nodes().Get(context.Background(), testNodeName, metav1.GetOptions{})
	assert.False(node.Spec.Unschedulable, "Node should be schedulable")
}
