// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"

	kubeupgradev1alpha1 "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
	applyconfigurationkubeupgradev1alpha1 "github.com/heathcliff26/kube-upgrade/pkg/client/applyconfiguration/kubeupgrade/v1alpha1"
	scheme "github.com/heathcliff26/kube-upgrade/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// KubeUpgradePlansGetter has a method to return a KubeUpgradePlanInterface.
// A group's client should implement this interface.
type KubeUpgradePlansGetter interface {
	KubeUpgradePlans(namespace string) KubeUpgradePlanInterface
}

// KubeUpgradePlanInterface has methods to work with KubeUpgradePlan resources.
type KubeUpgradePlanInterface interface {
	Create(ctx context.Context, kubeUpgradePlan *kubeupgradev1alpha1.KubeUpgradePlan, opts v1.CreateOptions) (*kubeupgradev1alpha1.KubeUpgradePlan, error)
	Update(ctx context.Context, kubeUpgradePlan *kubeupgradev1alpha1.KubeUpgradePlan, opts v1.UpdateOptions) (*kubeupgradev1alpha1.KubeUpgradePlan, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, kubeUpgradePlan *kubeupgradev1alpha1.KubeUpgradePlan, opts v1.UpdateOptions) (*kubeupgradev1alpha1.KubeUpgradePlan, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*kubeupgradev1alpha1.KubeUpgradePlan, error)
	List(ctx context.Context, opts v1.ListOptions) (*kubeupgradev1alpha1.KubeUpgradePlanList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *kubeupgradev1alpha1.KubeUpgradePlan, err error)
	Apply(ctx context.Context, kubeUpgradePlan *applyconfigurationkubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration, opts v1.ApplyOptions) (result *kubeupgradev1alpha1.KubeUpgradePlan, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, kubeUpgradePlan *applyconfigurationkubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration, opts v1.ApplyOptions) (result *kubeupgradev1alpha1.KubeUpgradePlan, err error)
	KubeUpgradePlanExpansion
}

// kubeUpgradePlans implements KubeUpgradePlanInterface
type kubeUpgradePlans struct {
	*gentype.ClientWithListAndApply[*kubeupgradev1alpha1.KubeUpgradePlan, *kubeupgradev1alpha1.KubeUpgradePlanList, *applyconfigurationkubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration]
}

// newKubeUpgradePlans returns a KubeUpgradePlans
func newKubeUpgradePlans(c *KubeupgradeV1alpha1Client, namespace string) *kubeUpgradePlans {
	return &kubeUpgradePlans{
		gentype.NewClientWithListAndApply[*kubeupgradev1alpha1.KubeUpgradePlan, *kubeupgradev1alpha1.KubeUpgradePlanList, *applyconfigurationkubeupgradev1alpha1.KubeUpgradePlanApplyConfiguration](
			"kubeupgradeplans",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *kubeupgradev1alpha1.KubeUpgradePlan { return &kubeupgradev1alpha1.KubeUpgradePlan{} },
			func() *kubeupgradev1alpha1.KubeUpgradePlanList { return &kubeupgradev1alpha1.KubeUpgradePlanList{} },
		),
	}
}