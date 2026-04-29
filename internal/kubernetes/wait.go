package kubernetes

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WaitForDelete waits for the provide runtime objects to be deleted from cluster
func WaitForDelete(c *RetryClient, objs []runtime.Object) error {
	// Wait for resources to be deleted.
	return wait.PollUntilContextTimeout(context.TODO(), 100*time.Millisecond, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		for _, obj := range objs {
			actual := &unstructured.Unstructured{}
			actual.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
			err = c.Get(ctx, ObjectKey(obj), actual)
			// Retry on transient API errors (return nil to keep polling) rather
			// than aborting the entire wait, which would surface as a misleading
			// "timed out" error well before the real deadline.
			if !errors.IsNotFound(err) {
				return false, nil
			}
		}

		return true, nil
	})
}

// WaitForSA waits for a service account to be present
func WaitForSA(config *rest.Config, name, namespace string) error {
	c, err := NewRetryClient(config, client.Options{
		Scheme: Scheme(),
	})
	if err != nil {
		return err
	}

	obj := &corev1.ServiceAccount{}

	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	return wait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, 60*time.Second, true, func(ctx context.Context) (done bool, err error) {
		// Retry on all errors (not-found, transient) rather than aborting the wait.
		return c.Get(ctx, key, obj) == nil, nil
	})
}
