package v1beta1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name            string
		testResourceRef TestResourceRef
		errored         bool
		expectedError   error
	}{
		{
			name: "apiVersion is not specified",
			testResourceRef: TestResourceRef{
				Kind:      "Pod",
				Namespace: "test",
				Name:      "test-pod",
				Ref:       "testPod",
			},
			errored:       true,
			expectedError: errAPIVersionInvalid,
		},
		{
			name: "apiVersion is invalid",
			testResourceRef: TestResourceRef{
				APIVersion: "x/y/z",
				Kind:       "Pod",
				Namespace:  "test",
				Name:       "test-pod",
				Ref:        "testPod",
			},
			errored:       true,
			expectedError: errAPIVersionInvalid,
		},
		{
			name: "apiVersion is valid and group is vacuous",
			testResourceRef: TestResourceRef{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  "test",
				Name:       "test-pod",
				Ref:        "testPod",
			},
			errored: false,
		},
		{
			name: "apiVersion has both group name and version",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "test",
				Name:       "test-deployment",
				Ref:        "testDeployment",
			},
			errored: false,
		},
		{
			name: "kind is not specified",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Namespace:  "test",
				Name:       "test-deployment",
				Ref:        "testDeployment",
			},
			errored:       true,
			expectedError: errKindNotSpecified,
		},
		{
			name: "name is not specified",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "test",
				Ref:        "testDeployment",
			},
			errored:       true,
			expectedError: errNameNotSpecified,
		},
		{
			name: "ref is not specified",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "test",
				Name:       "test-deployment",
			},
			errored:       true,
			expectedError: errRefNotSpecified,
		},
		{
			name: "all attributes are present and valid",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "test",
				Name:       "test-deployment",
				Ref:        "testDeployment",
			},
			errored: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testResourceRef.Validate()
			if !tc.errored {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectedError)
			}
		})
	}
}

func TestBuildResourceReference(t *testing.T) {
	buildObject := func(gvk schema.GroupVersionKind) *unstructured.Unstructured {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		return obj
	}

	testCases := []struct {
		name              string
		testResourceRef   TestResourceRef
		namespacedName    types.NamespacedName
		resourceReference *unstructured.Unstructured
	}{
		{
			name: "group name is vacuous",
			testResourceRef: TestResourceRef{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  "test",
				Name:       "test-pod",
				Ref:        "testPod",
			},
			namespacedName: types.NamespacedName{
				Namespace: "test",
				Name:      "test-pod",
			},
			resourceReference: buildObject(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}),
		},
		{
			name: "group name is present",
			testResourceRef: TestResourceRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "test",
				Name:       "test-deployment",
				Ref:        "testDeployment",
			},
			namespacedName: types.NamespacedName{
				Namespace: "test",
				Name:      "test-deployment",
			},
			resourceReference: buildObject(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namspacedName, referencedResource := tc.testResourceRef.BuildResourceReference()
			assert.Equal(t, tc.namespacedName, namspacedName)
			assert.True(
				t,
				reflect.DeepEqual(tc.resourceReference, referencedResource),
				"constructed unstructured reference does not match, expected '%s', got '%s'",
				tc.resourceReference,
				referencedResource,
			)
		})
	}
}
