package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func testObj() runtime.Object {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	obj.SetName("cm")
	obj.SetNamespace("default")
	return obj
}

func TestWaitForDelete_AlreadyGone(t *testing.T) {
	cl := fake.NewClientBuilder().Build()
	rc := &RetryClient{Client: cl}

	err := WaitForDelete(rc, []runtime.Object{testObj()})
	require.NoError(t, err)
}

func TestWaitForDelete_TransientErrorThenGone(t *testing.T) {
	callCount := 0
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			callCount++
			if callCount <= 3 {
				return fmt.Errorf("transient API error")
			}
			return k8serrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "cm")
		},
	}).Build()
	rc := &RetryClient{Client: cl}

	err := WaitForDelete(rc, []runtime.Object{testObj()})
	require.NoError(t, err)
	assert.Greater(t, callCount, 3)
}

func TestWaitForDelete_StillExistsThenGone(t *testing.T) {
	callCount := 0
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			callCount++
			if callCount <= 2 {
				return nil
			}
			return k8serrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "cm")
		},
	}).Build()
	rc := &RetryClient{Client: cl}

	err := WaitForDelete(rc, []runtime.Object{testObj()})
	require.NoError(t, err)
	assert.Greater(t, callCount, 2)
}

func TestWaitForDelete_PersistentErrorTimesOut(t *testing.T) {
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			return fmt.Errorf("persistent API error")
		},
	}).Build()
	rc := &RetryClient{Client: cl}

	err := WaitForDelete(rc, []runtime.Object{testObj()})
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
