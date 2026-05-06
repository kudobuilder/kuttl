package kubernetes

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WaitForDelete waits for the provided runtime objects to be deleted from cluster, up to duration.
// Retries on transient errors.
func WaitForDelete(cl client.Client, toDelete []client.Object, duration time.Duration) error {
	lastCheckMsg := ""
	err := wait.PollUntilContextTimeout(context.TODO(), 100*time.Millisecond, duration, true, func(ctx context.Context) (done bool, err error) {
		for _, obj := range toDelete {
			actual := &unstructured.Unstructured{}
			actual.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
			err = cl.Get(ctx, ObjectKey(obj), actual)
			if err == nil {
				lastCheckMsg = fmt.Sprintf("%v %s still exists", obj.GetObjectKind().GroupVersionKind(), obj.GetName())
				return false, nil
			}
			if !errors.IsNotFound(err) {
				lastCheckMsg = fmt.Sprintf("checking existence of %v %s failed: %v", obj.GetObjectKind().GroupVersionKind(), obj.GetName(), err)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for resource deletion (result of last check was: %q): %w", lastCheckMsg, err)
	}
	return nil
}
