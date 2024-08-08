package lock

import (
	"context"
	"testing"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/utils"
	"github.com/stretchr/testify/assert"
	coordv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testName      = "test-pod"
	testNamespace = "locktest-ns"
	testGroup     = "nodes"
	testNode      = "node1"
)

func TestNewLockUtility(t *testing.T) {
	assert := assert.New(t)

	l, err := NewLockUtility(nil, testName, testNamespace)
	assert.Error(err, "Should throw error")
	assert.Nil(l, "Should not return a lock utility")

	l, err = NewLockUtility(fake.NewSimpleClientset(), "", testNamespace)
	assert.Error(err, "Should throw error")
	assert.Nil(l, "Should not return a lock utility")

	l, err = NewLockUtility(fake.NewSimpleClientset(), testName, "")
	assert.Error(err, "Should throw error")
	assert.Nil(l, "Should not return a lock utility")

	l, err = NewLockUtility(fake.NewSimpleClientset(), testName, testNamespace)
	assert.NoError(err, "Should not throw error")
	assert.NotNil(l, "Should return a lock utility")
}

func TestLock(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		ok, err := l.Lock(testGroup, testNode)

		if !assert.NoError(err, "Should not return an error") {
			t.FailNow()
		}
		assert.True(ok, "Should succeed")

		lease, err := client.CoordinationV1().Leases(testNamespace).Get(context.Background(), leaseName(testGroup), metav1.GetOptions{})
		if !assert.NoError(err, "Should have the lease") {
			t.FailNow()
		}
		assert.Equal(utils.Pointer(testName), lease.Spec.HolderIdentity, "Lease should be hold by app")
		assert.Equal(utils.Pointer(defaultLeaseDurationSeconds), lease.Spec.LeaseDurationSeconds, "LeaseDurationSeconds should be set")
		assert.Equal(time.Now().Round(time.Second*5), lease.Spec.AcquireTime.Time.Round(time.Second*5), "AcquireTime should be now")
		assert.Equal(time.Now().Round(time.Second*5), lease.Spec.RenewTime.Time.Round(time.Second*5), "RenewTime should be now")
		assert.Equal(map[string]string{"node": testNode}, lease.ObjectMeta.Annotations, "Should have node annotation set")
	})
	t.Run("ValidLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": testNode},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("not-the-holder"),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
				RenewTime:            &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.Lock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.False(ok, "Should not aquire lease")
	})
	t.Run("AlreadyHasLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": testNode},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(l.name),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
				RenewTime:            &metav1.MicroTime{Time: time.Now().Add(-4 * time.Minute)},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.Lock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.True(ok, "Should succeed")

		lease, err = client.CoordinationV1().Leases(testNamespace).Get(context.Background(), leaseName(testGroup), metav1.GetOptions{})
		if !assert.NoError(err, "Should have the lease") {
			t.FailNow()
		}
		assert.Equal(time.Now().Round(time.Second*5), lease.Spec.RenewTime.Time.Round(time.Second*5), "Should have updated RenewTime")
	})
	t.Run("ExpiredLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": testNode},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("not-the-app"),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
				RenewTime:            &metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.Lock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.True(ok, "Should succeed")

		lease, err = client.CoordinationV1().Leases(testNamespace).Get(context.Background(), leaseName(testGroup), metav1.GetOptions{})
		if !assert.NoError(err, "Should have the lease") {
			t.FailNow()
		}
		assert.Equal(utils.Pointer(testName), lease.Spec.HolderIdentity, "Lease should be hold by app")
		assert.Equal(time.Now().Round(time.Second*5), lease.Spec.AcquireTime.Time.Round(time.Second*5), "AcquireTime should be now")
		assert.Equal(time.Now().Round(time.Second*5), lease.Spec.RenewTime.Time.Round(time.Second*5), "RenewTime should be now")
	})
	t.Run("ExpiredLeaseWrongNode", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": "different-node"},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("not-the-holder"),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
				RenewTime:            &metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.Lock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.False(ok, "Should not aquire lease")
	})
}

func TestRelease(t *testing.T) {
	client := fake.NewSimpleClientset()

	assert := assert.New(t)

	l, err := NewLockUtility(client, testName, testNamespace)
	if !assert.NoError(err, "Should create lock utility") {
		t.FailNow()
	}

	err = l.Release(testGroup)
	assert.NoError(err, "Should not return an error when no lease exists")

	err = l.createLease(testGroup, testNode)
	if !assert.NoError(err, "Should create lease") {
		t.FailNow()
	}

	err = l.Release(testGroup)
	assert.NoError(err, "Should release existing lease")

	_, err = client.CoordinationV1().Leases(testNamespace).Get(context.Background(), leaseName(testNode), metav1.GetOptions{})
	assert.True(errors.IsNotFound(err), "Lease should be deleted")
}

func TestHasLock(t *testing.T) {
	t.Run("NoLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		ok, err := l.HasLock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.False(ok, "Should not have lock")
	})
	t.Run("HasLease", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}
		err = l.createLease(testGroup, testNode)
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.HasLock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.True(ok, "Should have lock")
	})
	t.Run("WrongHolder", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": testNode},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("not-the-holder"),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
				RenewTime:            &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.HasLock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.False(ok, "Should not have lock")
	})
	t.Run("WrongNode", func(t *testing.T) {
		client := fake.NewSimpleClientset()

		assert := assert.New(t)

		l, err := NewLockUtility(client, testName, testNamespace)
		if !assert.NoError(err, "Should create lock utility") {
			t.FailNow()
		}

		lease := &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   l.namespace,
				Name:        leaseName(testGroup),
				Annotations: map[string]string{"node": "wrong-node"},
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer(testName),
				LeaseDurationSeconds: utils.Pointer(defaultLeaseDurationSeconds),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
				RenewTime:            &metav1.MicroTime{Time: time.Now()},
			},
		}
		_, err = client.CoordinationV1().Leases(l.namespace).Create(context.Background(), lease, metav1.CreateOptions{})
		if !assert.NoError(err, "Should create lease") {
			t.FailNow()
		}

		ok, err := l.HasLock(testGroup, testNode)
		assert.NoError(err, "Should not return an error")
		assert.False(ok, "Should not have lock")
	})
}
