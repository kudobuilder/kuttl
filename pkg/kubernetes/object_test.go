package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kudobuilder/kuttl/pkg/kubernetes/fake"
)

func TestGETAPIResource(t *testing.T) {
	fakeClient := fake.DiscoveryClient()

	apiResource, err := GetAPIResource(fakeClient, schema.GroupVersionKind{
		Kind:    "Pod",
		Version: "v1",
	})
	assert.Nil(t, err)
	assert.Equal(t, apiResource.Kind, "Pod")

	_, err = GetAPIResource(fakeClient, schema.GroupVersionKind{
		Kind:    "NonExistentResourceType",
		Version: "v1",
	})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "resource type not found")
}

func TestNamespaced(t *testing.T) {
	fakeClient := fake.DiscoveryClient()

	for _, test := range []struct {
		testName    string
		resource    runtime.Object
		namespace   string
		shouldError bool
	}{
		{
			testName:  "namespaced resource",
			resource:  NewPod("hello", ""),
			namespace: "set-the-namespace",
		},
		{
			testName:  "namespace already set",
			resource:  NewPod("hello", "other"),
			namespace: "other",
		},
		{
			testName:  "not-namespaced resource",
			resource:  NewResource("v1", "Namespace", "hello", ""),
			namespace: "",
		},
		{
			testName:    "non-existent resource",
			resource:    NewResource("v1", "Blah", "hello", ""),
			shouldError: true,
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			m, err := meta.Accessor(test.resource)
			require.NoError(t, err)

			actualName, actualNamespace, err := Namespaced(fakeClient, test.resource, "set-the-namespace")

			if test.shouldError {
				assert.NotNil(t, err)
				assert.Equal(t, "", actualName)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, m.GetName(), actualName)
			}

			assert.Equal(t, test.namespace, actualNamespace)
			assert.Equal(t, test.namespace, m.GetNamespace())
		})
	}
}
