package fake

import (
	v13 "k8s.io/api/apps/v1"
	v14 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v12 "k8s.io/api/core/v1"
	v15 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// DiscoveryClient returns a fake discovery client that is populated with some types for use in
// unit tests.
func DiscoveryClient() discovery.DiscoveryInterface {
	return &fake.FakeDiscovery{
		Fake: &testing.Fake{
			Resources: []*v1.APIResourceList{
				{
					GroupVersion: v12.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "pod", Namespaced: true, Kind: "Pod"},
						{Name: "namespace", Namespaced: false, Kind: "Namespace"},
						{Name: "service", Namespaced: true, Kind: "Service"},
					},
				},
				{
					GroupVersion: v13.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "statefulset", Namespaced: true, Kind: "StatefulSet"},
						{Name: "deployment", Namespaced: true, Kind: "Deployment"},
					},
				},
				{
					GroupVersion: v14.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "job", Namespaced: true, Kind: "Job"},
					},
				},
				{
					GroupVersion: v1beta1.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "job", Namespaced: true, Kind: "CronJob"},
					},
				},
				{
					GroupVersion: v15.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition"},
					},
				},
				{
					GroupVersion: v1beta12.SchemeGroupVersion.String(),
					APIResources: []v1.APIResource{
						{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition"},
					},
				},
			},
		},
	}
}
