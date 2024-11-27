package v1beta1

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (t *TestResourceRef) BuildResourceReference() (namespacedName types.NamespacedName, referencedResource *unstructured.Unstructured) {
	referencedResource = &unstructured.Unstructured{}
	apiVersionSplit := strings.Split(t.APIVersion, "/")
	gvk := schema.GroupVersionKind{
		Version: apiVersionSplit[len(apiVersionSplit)-1],
		Kind:    t.Kind,
	}
	if len(t.APIVersion) > 1 {
		gvk.Group = apiVersionSplit[0]
	}
	referencedResource.SetGroupVersionKind(gvk)

	namespacedName = types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Name,
	}

	return
}

func (t *TestResourceRef) ValidateResourceReference() error {
	apiVersionSplit := strings.Split(t.APIVersion, "/")
	if t.APIVersion == "" || (len(apiVersionSplit) != 1 && len(apiVersionSplit) != 2) {
		return fmt.Errorf("apiVersion '%v' not of the format (<group>/)<version>", t.APIVersion)
	} else if t.Kind == "" {
		return errors.New("kind not specified")
	} else if t.Namespace == "" {
		return errors.New("namespace not specified")
	} else if t.Name == "" {
		return errors.New("name not specified")
	} else if t.Ref == "" {
		return errors.New("ref not specified")
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
