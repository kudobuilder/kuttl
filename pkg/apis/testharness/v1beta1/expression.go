package v1beta1

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	apiVersionInvalidErr = errors.New("apiVersion not of the format (<group>/)<version>")
	kindNotSpecifiedErr  = errors.New("kind not specified")
	nameNotSpecifiedErr  = errors.New("name not specified")
	refNotSpecifiedErr   = errors.New("ref not specified")
)

func (t *TestResourceRef) BuildResourceReference() (namespacedName types.NamespacedName, referencedResource *unstructured.Unstructured) {
	referencedResource = &unstructured.Unstructured{}
	apiVersionSplit := strings.Split(t.APIVersion, "/")
	gvk := schema.GroupVersionKind{
		Version: apiVersionSplit[len(apiVersionSplit)-1],
		Kind:    t.Kind,
	}
	if len(apiVersionSplit) > 1 {
		gvk.Group = apiVersionSplit[0]
	}
	referencedResource.SetGroupVersionKind(gvk)

	namespacedName = types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Name,
	}

	return
}

func (t *TestResourceRef) Validate() error {
	apiVersionSplit := strings.Split(t.APIVersion, "/")
	switch {
	case t.APIVersion == "" || (len(apiVersionSplit) != 1 && len(apiVersionSplit) != 2):
		return apiVersionInvalidErr
	case t.Kind == "":
		return kindNotSpecifiedErr
	case t.Name == "":
		return nameNotSpecifiedErr
	case t.Ref == "":
		return refNotSpecifiedErr
	}

	return nil
}

func (t *TestResourceRef) String() string {
	return fmt.Sprintf(
		"apiVersion=%v, kind=%v, namespace=%v, name=%v, ref=%v",
		t.APIVersion,
		t.Kind,
		t.Namespace,
		t.Name,
		t.Ref,
	)
}
