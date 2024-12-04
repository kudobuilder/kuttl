package kubernetes

import (
	"fmt"

	v13 "k8s.io/api/apps/v1"
	v14 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	v15 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kudobuilder/kuttl/pkg/test/utils"
)

var ImpersonateAs = ""

func GetConfig() (*rest.Config, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	if ImpersonateAs != "" {
		cfg.Impersonate.UserName = ImpersonateAs
	}

	return cfg, nil
}

func BuildConfigWithContext(kubeconfig, context string) (*rest.Config, error) {
	if context == "" {
		// Use default context
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
}

// ResourceID returns a human readable identifier indicating the object kind, name, and namespace.
func ResourceID(obj runtime.Object) string {
	m, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}

	gvk := obj.GetObjectKind().GroupVersionKind()

	return fmt.Sprintf("%s:%s/%s", gvk.Kind, m.GetNamespace(), m.GetName())
}

// Namespaced sets the namespace on an object to namespace, if it is a namespace scoped resource.
// If the resource is cluster scoped, then it is ignored and the namespace is not set.
// If it is a namespaced resource and a namespace is already set, then the namespace is unchanged.
func Namespaced(dClient discovery.DiscoveryInterface, obj runtime.Object, namespace string) (string, string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return "", "", err
	}

	if m.GetNamespace() != "" {
		return m.GetName(), m.GetNamespace(), nil
	}

	resource, err := utils.GetAPIResource(dClient, obj.GetObjectKind().GroupVersionKind())
	if err != nil {
		return "", "", fmt.Errorf("retrieving API resource for %v failed: %v", obj.GetObjectKind().GroupVersionKind(), err)
	}

	if !resource.Namespaced {
		return m.GetName(), "", nil
	}

	m.SetNamespace(namespace)
	return m.GetName(), namespace, nil
}

// MatchesKind returns true if the Kubernetes kind of obj matches any of kinds.
func MatchesKind(obj runtime.Object, kinds ...runtime.Object) bool {
	gvk := obj.GetObjectKind().GroupVersionKind()

	for _, kind := range kinds {
		if kind.GetObjectKind().GroupVersionKind() == gvk {
			return true
		}
	}

	return false
}

// ObjectKey returns an instantiated ObjectKey for the provided object.
func ObjectKey(obj runtime.Object) client.ObjectKey {
	m, _ := meta.Accessor(obj) //nolint:errcheck // runtime.Object don't have the error issues of interface{}
	return client.ObjectKey{
		Name:      m.GetName(),
		Namespace: m.GetNamespace(),
	}
}

// NewV1Pod returns a new corev1.Pod object.
// Each of name, namespace and serviceAccountName are set if non-empty.
func NewV1Pod(name, namespace, serviceAccountName string) *v1.Pod {
	pod := v1.Pod{}
	pod.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	})
	if name != "" {
		pod.SetName(name)
	}
	if namespace != "" {
		pod.SetNamespace(namespace)
	}
	if serviceAccountName != "" {
		pod.Spec.ServiceAccountName = serviceAccountName
	}
	return &pod
}

// FakeDiscoveryClient returns a fake discovery client that is populated with some types for use in
// unit tests.
func FakeDiscoveryClient() discovery.DiscoveryInterface {
	return &fake.FakeDiscovery{
		Fake: &testing.Fake{
			Resources: []*v12.APIResourceList{
				{
					GroupVersion: v1.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "pod", Namespaced: true, Kind: "Pod"},
						{Name: "namespace", Namespaced: false, Kind: "Namespace"},
						{Name: "service", Namespaced: true, Kind: "Service"},
					},
				},
				{
					GroupVersion: v13.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "statefulset", Namespaced: true, Kind: "StatefulSet"},
						{Name: "deployment", Namespaced: true, Kind: "Deployment"},
					},
				},
				{
					GroupVersion: v14.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "job", Namespaced: true, Kind: "Job"},
					},
				},
				{
					GroupVersion: v1beta1.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "job", Namespaced: true, Kind: "CronJob"},
					},
				},
				{
					GroupVersion: v15.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition"},
					},
				},
				{
					GroupVersion: v1beta12.SchemeGroupVersion.String(),
					APIResources: []v12.APIResource{
						{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition"},
					},
				},
			},
		},
	}
}
