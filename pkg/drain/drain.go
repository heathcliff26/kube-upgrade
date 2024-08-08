package drain

import (
	"context"
	"fmt"
	"log/slog"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type Drainer struct {
	client kubernetes.Interface
}

// Create a new drain utility
func NewDrainer(client kubernetes.Interface) (*Drainer, error) {
	if client == nil {
		return nil, fmt.Errorf("failed to create drain utility: no kubernetes client provided")
	}

	return &Drainer{
		client: client,
	}, nil
}

// Drain a node of all pods, skipping daemonsets
func (d *Drainer) DrainNode(node string) error {
	_, err := d.client.CoreV1().Nodes().Patch(context.Background(), node, types.MergePatchType, nodeUnschedulablePatch(true), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	pods, err := d.client.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	})
	if err != nil {
		return err
	}

	var returnError error
	for _, pod := range pods.Items {
		// Skip mirror pods
		if _, ok := pod.ObjectMeta.Annotations[v1.MirrorPodAnnotationKey]; ok {
			continue
		}
		// Skip daemonsets
		controller := metav1.GetControllerOf(&pod)
		if controller != nil && controller.Kind == "DaemonSet" {
			continue
		}

		err = d.client.PolicyV1().Evictions(pod.GetNamespace()).Evict(context.Background(), &policyv1.Eviction{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "policy/v1",
				Kind:       "Eviction",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			},
			DeleteOptions: metav1.NewDeleteOptions(*pod.Spec.TerminationGracePeriodSeconds),
		})
		if err != nil {
			slog.Info("Failed to evict pod", "err", err, slog.String("node", node), slog.String("pod", pod.GetName()), slog.String("namespace", pod.GetNamespace()))
			if returnError == nil {
				returnError = err
			}
			continue
		}
		slog.Info("Evicted pod", slog.String("node", node), slog.String("pod", pod.GetName()), slog.String("namespace", pod.GetNamespace()))
	}

	return returnError
}

// Uncordon a node
func (d *Drainer) UncordonNode(node string) error {
	_, err := d.client.CoreV1().Nodes().Patch(context.Background(), node, types.MergePatchType, nodeUnschedulablePatch(false), metav1.PatchOptions{})
	return err
}
