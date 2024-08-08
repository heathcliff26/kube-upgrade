package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/utils"
	coordv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coordclient "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

const defaultLeaseDurationSeconds = int32(300)

type LockUtility struct {
	client coordclient.LeasesGetter

	// Either the name of the app, pod or some other identifier
	name string
	// The namespace used for coordination, should be the same the current pod is running in
	namespace string
}

// Create a new lock utility
func NewLockUtility(client kubernetes.Interface, name, namespace string) (*LockUtility, error) {
	if client == nil {
		return nil, fmt.Errorf("failed to create lock utility: no kubernetes client provided")
	}
	if name == "" {
		return nil, fmt.Errorf("failed to create lock utility: no unique name provided")
	}
	if namespace == "" {
		return nil, fmt.Errorf("failed to create lock utility: no namespace provided")
	}

	return &LockUtility{
		client:    client.CoordinationV1(),
		name:      name,
		namespace: namespace,
	}, nil
}

// Lock a group to work on a specific node.
// Returns true if the current instance aquired or already has the lock.
func (l *LockUtility) Lock(group, node string) (bool, error) {
	lease, err := l.client.Leases(l.namespace).Get(context.Background(), leaseName(group), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		err := l.createLease(group, node)
		return err == nil, err
	} else if err != nil {
		return false, err
	}
	if lease.Spec.HolderIdentity != nil && *lease.Spec.HolderIdentity == l.name {
		err := l.renewLease(lease)
		return err == nil, err
	}

	// Check if the lease is expired
	if lease.Spec.RenewTime != nil &&
		lease.Spec.LeaseDurationSeconds != nil &&
		time.Now().After(lease.Spec.RenewTime.Time.Add(time.Duration(*lease.Spec.LeaseDurationSeconds)*time.Second)) {

		// Do not aquire leases for another node
		if lease.ObjectMeta.Annotations == nil || lease.ObjectMeta.Annotations["node"] != node {
			return false, nil
		}

		lease.Spec.HolderIdentity = utils.Pointer(l.name)
		lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now()}
		err := l.renewLease(lease)
		return err == nil, err
	}
	return false, nil
}

func (l *LockUtility) createLease(group, node string) error {
	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   l.namespace,
			Name:        leaseName(group),
			Annotations: map[string]string{"node": node},
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       utils.Pointer(l.name),
			LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			RenewTime:            &metav1.MicroTime{Time: time.Now()},
		},
	}

	_, err := l.client.Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
	return err
}

func (l *LockUtility) renewLease(lease *coordv1.Lease) error {
	lease.Spec.RenewTime = &metav1.MicroTime{Time: time.Now()}
	_, err := l.client.Leases(l.namespace).Update(context.Background(), lease, metav1.UpdateOptions{})
	return err
}

// Release a lock
func (l *LockUtility) Release(group string) error {
	err := l.client.Leases(l.namespace).Delete(context.Background(), leaseName(group), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

// Check if the lock for this node is hold
func (l *LockUtility) HasLock(group, node string) (bool, error) {
	lease, err := l.client.Leases(l.namespace).Get(context.Background(), leaseName(group), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if lease.Spec.HolderIdentity == nil || *lease.Spec.HolderIdentity != l.name {
		return false, nil
	}
	if lease.ObjectMeta.Annotations == nil || lease.ObjectMeta.Annotations["node"] != node {
		return false, nil
	}
	return true, nil
}
