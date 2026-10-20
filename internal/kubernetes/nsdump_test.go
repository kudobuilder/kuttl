package kubernetes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	discoveryfake "k8s.io/client-go/discovery/fake"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// capturingClient implements LoggingClient, recording all lines for assertions.
type capturingClient struct {
	client.Client

	logs []string
}

func (c *capturingClient) Logf(format string, args ...any) {
	c.logs = append(c.logs, fmt.Sprintf(format, args...))
}

func (c *capturingClient) String() string {
	return strings.Join(c.logs, "\n")
}

func newCapturingClient(objs ...client.Object) *capturingClient {
	return &capturingClient{
		Client: fake.NewClientBuilder().
			WithScheme(k8sscheme.Scheme).
			WithObjects(objs...).
			Build(),
	}
}

// preferredResourcesFake overrides ServerPreferredNamespacedResources because
// client-go's FakeDiscovery stubs it out, always returning nothing.
type preferredResourcesFake struct {
	*discoveryfake.FakeDiscovery
}

func (f *preferredResourcesFake) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return f.Resources, nil
}

// fakeDiscovery exposes a catalog where configmaps and services are listable
// while bindings are not, covering the verbs filter.
func fakeDiscovery() discovery.DiscoveryInterface {
	return &preferredResourcesFake{
		FakeDiscovery: &discoveryfake.FakeDiscovery{
			Fake: &k8stesting.Fake{
				Resources: []*metav1.APIResourceList{
					{
						GroupVersion: corev1.SchemeGroupVersion.String(),
						APIResources: []metav1.APIResource{
							{Name: "configmaps", Namespaced: true, Kind: "ConfigMap", Verbs: []string{"list"}},
							{Name: "services", Namespaced: true, Kind: "Service", Verbs: []string{"list"}},
							{Name: "bindings", Namespaced: true, Kind: "Binding", Verbs: []string{"create"}},
						},
					},
				},
			},
		},
	}
}

func stuckNamespace(name string) *corev1.Namespace {
	now := metav1.Now()
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Finalizers:        []string{"example.com/ns-stuck"},
			DeletionTimestamp: &now,
		},
		Status: corev1.NamespaceStatus{Phase: corev1.NamespaceTerminating},
	}
}

func stuckService(ns string) *corev1.Service {
	now := metav1.Now()
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         ns,
			Name:              "junk-service",
			Finalizers:        []string{"example.com/junk"},
			DeletionTimestamp: &now,
		},
	}
}

func TestDumpStuckNamespace(t *testing.T) {
	cl := newCapturingClient(stuckNamespace("stuck-ns"), stuckService("stuck-ns"))

	DumpStuckNamespace(cl, fakeDiscovery(), "stuck-ns")

	out := cl.String()
	assert.Contains(t, out, `stuck namespace "stuck-ns"`)
	assert.Contains(t, out, "example.com/ns-stuck")
	assert.Contains(t, out, "v1/Service/junk-service")
	assert.Contains(t, out, "example.com/junk")
	assert.NotContains(t, out, "Binding")
}

func TestDumpStuckNamespaceWhenNamespaceGone(t *testing.T) {
	cl := newCapturingClient()

	assert.NotPanics(t, func() {
		DumpStuckNamespace(cl, fakeDiscovery(), "gone-ns")
	})
	assert.Contains(t, cl.String(), `failed to get namespace "gone-ns"`)
}
