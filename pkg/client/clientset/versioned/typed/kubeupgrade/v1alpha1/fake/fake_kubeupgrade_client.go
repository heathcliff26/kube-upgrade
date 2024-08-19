// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/heathcliff26/kube-upgrade/pkg/client/clientset/versioned/typed/kubeupgrade/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeKubeupgradeV1alpha1 struct {
	*testing.Fake
}

func (c *FakeKubeupgradeV1alpha1) KubeUpgradePlans(namespace string) v1alpha1.KubeUpgradePlanInterface {
	return &FakeKubeUpgradePlans{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeKubeupgradeV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
