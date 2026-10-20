package testcase

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	discoveryfake "k8s.io/client-go/discovery/fake"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	testutils "github.com/kudobuilder/kuttl/internal/utils"
)

// capturingLogger implements testutils.Logger recording all lines for assertions.
type capturingLogger struct {
	logs []string
}

func (l *capturingLogger) Log(args ...any) {
	l.logs = append(l.logs, fmt.Sprint(args...))
}

func (l *capturingLogger) Logf(format string, args ...any) {
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}

func (l *capturingLogger) WithPrefix(_ string) testutils.Logger {
	return l
}

func (l *capturingLogger) Write(p []byte) (n int, err error) {
	l.logs = append(l.logs, string(p))
	return len(p), nil
}

func (l *capturingLogger) Flush() {}

func (l *capturingLogger) String() string {
	return strings.Join(l.logs, "\n")
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

func newCaseWithNS(nsName string, timeout int) *Case {
	return &Case{
		ns:      &namespace{name: nsName},
		timeout: timeout,
		getDiscoveryClient: func() (discovery.DiscoveryInterface, error) {
			return fakeDiscovery(), nil
		},
	}
}

func newTestClient(logger *capturingLogger, objs ...client.Object) clientWithKubeConfig {
	fakeClient := fake.NewClientBuilder().
		WithScheme(k8sscheme.Scheme).
		WithObjects(objs...).
		Build()
	return clientWithKubeConfig{Client: fakeClient, logger: logger}
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

func TestDeleteNamespaceDumpsContentsOnlyWhenStuck(t *testing.T) {
	t.Run("stuck namespace is dumped", func(t *testing.T) {
		logger := &capturingLogger{}
		cl := newTestClient(logger, stuckNamespace("stuck-ns"), stuckService("stuck-ns"))
		c := newCaseWithNS("stuck-ns", 1)

		err := c.deleteNamespace(cl)

		require.Error(t, err)
		assert.Contains(t, logger.String(), "stuck-ns")
		assert.Contains(t, logger.String(), "junk-service")
		assert.Contains(t, logger.String(), "example.com/junk")
	})

	t.Run("successfully deleted namespace is not dumped", func(t *testing.T) {
		logger := &capturingLogger{}
		cleanNS := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "clean-ns"},
		}
		cl := newTestClient(logger, cleanNS)
		c := newCaseWithNS("clean-ns", 5)

		err := c.deleteNamespace(cl)

		require.NoError(t, err)
		assert.NotContains(t, logger.String(), "failed to get namespace")
	})
}
