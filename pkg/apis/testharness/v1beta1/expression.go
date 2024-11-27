package v1beta1

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (t *TestResourceRef) BuildResourceReference() (types.NamespacedName, *unstructured.Unstructured) {
	apiVersionSplit := strings.Split(t.APIVersion, "/")
	gvk := schema.GroupVersionKind{
		Version: apiVersionSplit[len(apiVersionSplit)-1],
		Kind:    t.Kind,
	}
	if len(t.APIVersion) > 1 {
		gvk.Group = apiVersionSplit[0]
	}

	referencedResource := &unstructured.Unstructured{}
	referencedResource.SetGroupVersionKind(gvk)

	namespacedName := types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Name,
	}

	return namespacedName, referencedResource
}
