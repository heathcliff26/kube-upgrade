package controller

import (
	"fmt"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	upgradedconfig "github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// Creates a new DaemonSet with the required metadata and no spec.
func (c *controller) NewEmptyUpgradedDaemonSet(plan, group string) appv1.DaemonSet {
	labels := upgradedLabels(plan, group)

	return appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("upgraded-%s", group),
			Namespace: c.namespace,
			Labels:    labels,
		},
	}
}

// Create a Spec template for an upgraded group.
// Caller should add node selector after creation.
// Used to override the DaemonSet on each reconciliation.
func (c *controller) NewUpgradedDaemonSetSpec(plan, group string) appv1.DaemonSetSpec {
	labels := upgradedLabels(plan, group)

	spec := appv1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:      "node-role.kubernetes.io/control-plane",
						Operator: corev1.TolerationOpExists,
						Effect:   corev1.TaintEffectNoSchedule,
					},
					{
						Key:      "node-role.kubernetes.io/master",
						Operator: corev1.TolerationOpExists,
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
				// Need to run with host PIDs for rpm-ostree to work.
				// Otherwise it won't see the caller process PID.
				HostPID: true,
				Containers: []corev1.Container{
					{
						Name:  "upgraded",
						Image: c.upgradedImage,
						Env: []corev1.EnvVar{
							{
								Name: "HOSTNAME",
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
	}

	attachVolumeMountHostPath(&spec, "host-run", "/run", "/run")
	attachVolumeMountHostPath(&spec, "rootfs", "/", "/host")
	// Contains certificates referenced by kubelet config.
	attachVolumeMountHostPath(&spec, "kubelet-pki", "/var/lib/kubelet/pki", "/var/lib/kubelet/pki")
	// TODO: This could be dropped in favour of detecting the node name via hostname in upgraded.
	attachVolumeMountHostPath(&spec, "machine-id", "/etc/machine-id", "/etc/machine-id")

	return spec
}

// Creates a new ConfigMap with the required metadata and no spec.
func (c *controller) NewEmptyUpgradedConfigMap(plan, group string) corev1.ConfigMap {
	labels := upgradedLabels(plan, group)

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("upgraded-%s", group),
			Namespace: c.namespace,
			Labels:    labels,
		},
	}
}

// Converts the given config to a string and attaches it to the given ConfigMap data.
func (c *controller) AttachUpgradedConfigMapData(cm *corev1.ConfigMap, cfg *api.UpgradedConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to convert upgraded config to yaml: %v", err)
	}

	cm.Data = map[string]string{upgradedconfig.DefaultConfigFile: string(data)}
	return nil
}

func attachVolumeMountHostPath(spec *appv1.DaemonSetSpec, name, hostPath, mountPath string) {
	spec.Template.Spec.Volumes = append(spec.Template.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: hostPath,
			},
		},
	})
	spec.Template.Spec.Containers[0].VolumeMounts = append(spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
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
