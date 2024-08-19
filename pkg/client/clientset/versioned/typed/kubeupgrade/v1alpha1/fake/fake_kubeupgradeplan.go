// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
	kubeupgradev1alpha1 "github.com/heathcliff26/kube-upgrade/pkg/client/applyconfiguration/kubeupgrade/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKubeUpgradePlans implements KubeUpgradePlanInterface
type FakeKubeUpgradePlans struct {
	Fake *FakeKubeupgradeV1alpha1
	ns   string
}

var kubeupgradeplansResource = v1alpha1.SchemeGroupVersion.WithResource("kubeupgradeplans")

var kubeupgradeplansKind = v1alpha1.SchemeGroupVersion.WithKind("KubeUpgradePlan")

// Get takes name of the kubeUpgradePlan, and returns the corresponding kubeUpgradePlan object, and an error if there is any.
func (c *FakeKubeUpgradePlans) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(kubeupgradeplansResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// List takes label and field selectors, and returns the list of KubeUpgradePlans that match those selectors.
func (c *FakeKubeUpgradePlans) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KubeUpgradePlanList, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlanList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(kubeupgradeplansResource, kubeupgradeplansKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KubeUpgradePlanList{ListMeta: obj.(*v1alpha1.KubeUpgradePlanList).ListMeta}
	for _, item := range obj.(*v1alpha1.KubeUpgradePlanList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kubeUpgradePlans.
func (c *FakeKubeUpgradePlans) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(kubeupgradeplansResource, c.ns, opts))

}

// Create takes the representation of a kubeUpgradePlan and creates it.  Returns the server's representation of the kubeUpgradePlan, and an error, if there is any.
func (c *FakeKubeUpgradePlans) Create(ctx context.Context, kubeUpgradePlan *v1alpha1.KubeUpgradePlan, opts v1.CreateOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(kubeupgradeplansResource, c.ns, kubeUpgradePlan, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// Update takes the representation of a kubeUpgradePlan and updates it. Returns the server's representation of the kubeUpgradePlan, and an error, if there is any.
func (c *FakeKubeUpgradePlans) Update(ctx context.Context, kubeUpgradePlan *v1alpha1.KubeUpgradePlan, opts v1.UpdateOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(kubeupgradeplansResource, c.ns, kubeUpgradePlan, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKubeUpgradePlans) UpdateStatus(ctx context.Context, kubeUpgradePlan *v1alpha1.KubeUpgradePlan, opts v1.UpdateOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(kubeupgradeplansResource, "status", c.ns, kubeUpgradePlan, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// Delete takes name of the kubeUpgradePlan and deletes it. Returns an error if one occurs.
func (c *FakeKubeUpgradePlans) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(kubeupgradeplansResource, c.ns, name, opts), &v1alpha1.KubeUpgradePlan{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKubeUpgradePlans) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(kubeupgradeplansResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.KubeUpgradePlanList{})
	return err
}

// Patch applies the patch and returns the patched kubeUpgradePlan.
func (c *FakeKubeUpgradePlans) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KubeUpgradePlan, err error) {
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(kubeupgradeplansResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied kubeUpgradePlan.
func (c *FakeKubeUpgradePlans) Apply(ctx context.Context, kubeUpgradePlan *kubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	if kubeUpgradePlan == nil {
		return nil, fmt.Errorf("kubeUpgradePlan provided to Apply must not be nil")
	}
	data, err := json.Marshal(kubeUpgradePlan)
	if err != nil {
		return nil, err
	}
	name := kubeUpgradePlan.Name
	if name == nil {
		return nil, fmt.Errorf("kubeUpgradePlan.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(kubeupgradeplansResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeKubeUpgradePlans) ApplyStatus(ctx context.Context, kubeUpgradePlan *kubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.KubeUpgradePlan, err error) {
	if kubeUpgradePlan == nil {
		return nil, fmt.Errorf("kubeUpgradePlan provided to Apply must not be nil")
	}
	data, err := json.Marshal(kubeUpgradePlan)
	if err != nil {
		return nil, err
	}
	name := kubeUpgradePlan.Name
	if name == nil {
		return nil, fmt.Errorf("kubeUpgradePlan.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.KubeUpgradePlan{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(kubeupgradeplansResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.KubeUpgradePlan), err
}
