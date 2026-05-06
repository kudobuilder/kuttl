package kubernetes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func testObj() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	obj.SetName("cm")
	obj.SetNamespace("default")
	return obj
}

func TestWaitForDelete_AlreadyGone(t *testing.T) {
	cl := fake.NewClientBuilder().Build()

	err := WaitForDelete(cl, []client.Object{testObj()}, time.Second*2)
	require.NoError(t, err)
}

func TestWaitForDelete_TransientErrorThenGone(t *testing.T) {
	callCount := 0
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			callCount++
			if callCount <= 3 {
				return fmt.Errorf("transient API error")
			}
			return k8serrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "cm")
		},
	}).Build()

	err := WaitForDelete(cl, []client.Object{testObj()}, time.Second*2)
	require.NoError(t, err)
	assert.Greater(t, callCount, 3)
}

func TestWaitForDelete_StillExistsThenGone(t *testing.T) {
	callCount := 0
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			callCount++
			if callCount <= 2 {
				return nil
			}
			return k8serrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "cm")
		},
	}).Build()

	err := WaitForDelete(cl, []client.Object{testObj()}, time.Second*2)
	require.NoError(t, err)
	assert.Greater(t, callCount, 2)
}

func TestWaitForDelete_PersistentErrorTimesOut(t *testing.T) {
	callCount := 0
	cl := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			callCount++
			if callCount <= 2 {
				return fmt.Errorf("initial transient error")
			}
			return fmt.Errorf("persistent API error")
		},
	}).Build()

	err := WaitForDelete(cl, []client.Object{testObj()}, time.Second*2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.ErrorContains(t, err, "result of last check was:")
	assert.ErrorContains(t, err, "failed: persistent API error")
	assert.NotContains(t, err.Error(), "initial transient error")
}
