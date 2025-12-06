package controller

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	upgradedconfig "github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

// Creates a new DaemonSet with the required metadata and spec.
// Caller should add node selector after creation.
func (c *controller) NewUpgradedDaemonSet(plan, group string) *appv1.DaemonSet {
	labels := upgradedLabels(plan, group)

	ds := &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("upgraded-%s", group),
			Namespace: c.namespace,
			Labels:    labels,
		},
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					// Need to run with host PIDs for rpm-ostree to work.
					// Otherwise it won't see the caller process PID.
					HostPID: true,
					Containers: []corev1.Container{
						{
							Name:  "upgraded",
							Image: c.upgradedImage,
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: Pointer(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: upgradedconfig.DefaultConfigDir,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("upgraded-%s", group),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	attachVolumeMountHostPath(ds, "host-run", "/run", "/run")
	attachVolumeMountHostPath(ds, "rootfs", "/", "/host")
	// Contains certificates referenced by kubelet config.
	attachVolumeMountHostPath(ds, "kubelet-pki", "/var/lib/kubelet/pki", "/var/lib/kubelet/pki")
	attachVolumeMountHostPath(ds, "machine-id", "/etc/machine-id", "/etc/machine-id")

	return ds
}

// Creates a new ConfigMap with the required metadata and data.
func (c *controller) NewUpgradedConfigMap(plan, group string, cfg *api.UpgradedConfig) (*corev1.ConfigMap, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert upgraded config to yaml: %v", err)
	}

	labels := upgradedLabels(plan, group)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("upgraded-%s", group),
			Namespace: c.namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			upgradedconfig.DefaultConfigFile: string(data),
		},
	}, nil
}

// Reconcile the given ConfigMap with the expected state from the given config.
func (c *controller) reconcileUpgradedConfigMap(ctx context.Context, plan *api.KubeUpgradePlan, logger *slog.Logger, cm *corev1.ConfigMap, group string) error {
	upgradedCfg := combineConfig(plan.Spec.Upgraded, plan.Spec.Groups[group].Upgraded)

	expectedCM, err := c.NewUpgradedConfigMap(plan.Name, group, upgradedCfg)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(plan, expectedCM, c.Scheme())
	if err != nil {
		return err
	}

	logger = logger.With("configmap", expectedCM.Name, "namespace", expectedCM.Namespace)

	if cm == nil {
		logger.Info("Creating upgraded ConfigMap for group")
		return c.Create(ctx, expectedCM)
	}

	updated := false

	if !reflect.DeepEqual(expectedCM.OwnerReferences, cm.OwnerReferences) {
		cm.OwnerReferences = expectedCM.OwnerReferences
		updated = true
		logger.Debug("Reconciling metadata.ownerReferences")
	}

	if !reflect.DeepEqual(expectedCM.Labels, cm.Labels) {
		cm.Labels = expectedCM.Labels
		updated = true
		logger.Debug("Reconciling metadata.labels")
	}

	if !reflect.DeepEqual(expectedCM.Data, cm.Data) {
		cm.Data = expectedCM.Data
		updated = true
		logger.Debug("Reconciling configmap data")
	}

	if updated {
		logger.Debug("Updating ConfigMap")
		return c.Update(ctx, cm)
	}
	return nil
}

// Reconcile the given DaemonSet with the expected spec.
func (c *controller) reconcileUpgradedDaemonSet(ctx context.Context, plan *api.KubeUpgradePlan, logger *slog.Logger, ds *appv1.DaemonSet, groupName string, group api.KubeUpgradePlanGroup) error {
	expectedDS := c.NewUpgradedDaemonSet(plan.Name, groupName)
	expectedDS.Spec.Template.Spec.NodeSelector = group.Labels
	expectedDS.Spec.Template.Spec.Tolerations = group.Tolerations
	err := controllerutil.SetControllerReference(plan, expectedDS, c.Scheme())
	if err != nil {
		return err
	}

	logger = logger.With("daemonset", expectedDS.Name, "namespace", expectedDS.Namespace)

	if ds == nil {
		logger.Info("Creating upgraded DaemonSet for group")
		ds = expectedDS.DeepCopy()
		err := c.Create(ctx, ds, client.DryRunAll)
		if err != nil {
			return err
		}
		hash, err := createHash(expectedDS.Spec, ds.Spec)
		if err != nil {
			return err
		}
		annotateUpgradedDaemonSetHash(ds, hash)
		return c.Create(ctx, ds)
	}

	updated := false

	if !reflect.DeepEqual(expectedDS.OwnerReferences, ds.OwnerReferences) {
		ds.OwnerReferences = expectedDS.OwnerReferences
		updated = true
		logger.Debug("Reconciling metadata.ownerReferences")
	}

	if !reflect.DeepEqual(expectedDS.Labels, ds.Labels) {
		ds.Labels = expectedDS.Labels
		updated = true
		logger.Debug("Reconciling metadata.labels")
	}

	hash, err := createHash(expectedDS.Spec, ds.Spec)
	if err != nil {
		return err
	}
	if ds.Annotations == nil || ds.Annotations[constants.ControllerResourceHash] != hash {
		ds.Spec = expectedDS.Spec
		updated = true
		logger.Debug("Reconciling spec")

		err = c.Update(ctx, ds, client.DryRunAll)
		if err != nil {
			return err
		}
		hash, err = createHash(expectedDS.Spec, ds.Spec)
		if err != nil {
			return err
		}
		annotateUpgradedDaemonSetHash(ds, hash)
	}

	if updated {
		logger.Debug("Updating Daemonset")
		return c.Update(ctx, ds)
	}
	return nil
}

func attachVolumeMountHostPath(ds *appv1.DaemonSet, name, hostPath, mountPath string) {
	ds.Spec.Template.Spec.Volumes = append(ds.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: hostPath,
			},
		},
	})
	ds.Spec.Template.Spec.Containers[0].VolumeMounts = append(ds.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	})
}

func upgradedLabels(planName, groupName string) map[string]string {
	return map[string]string{
		constants.LabelPlanName:  planName,
		constants.LabelNodeGroup: groupName,
	}
}

// Create a hash from a list of objects by serializing them together as a JSON array and hashing the result.
// All provided objects are marshaled as a single JSON array, so the hash depends on their order and content.
func createHash(obj ...interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object for hashing: %v", err)
	}
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:]), nil
}

// Annotate the given DaemonSet with the given hash.
func annotateUpgradedDaemonSetHash(ds *appv1.DaemonSet, hash string) {
	if ds.Annotations == nil {
		ds.Annotations = make(map[string]string)
	}
	ds.Annotations[constants.ControllerResourceHash] = hash
}
