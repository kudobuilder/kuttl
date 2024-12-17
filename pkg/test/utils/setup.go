package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/kudobuilder/kuttl/pkg/kubernetes"
)

// StartTestEnvironment is a wrapper for controller-runtime's envtest that creates a Kubernetes API server and etcd
// suitable for use in tests.
func StartTestEnvironment(attachControlPlaneOutput bool) (env TestEnvironment, err error) {
	env.Environment = &envtest.Environment{
		AttachControlPlaneOutput: attachControlPlaneOutput,
	}

	env.Config, err = env.Environment.Start()

	if err != nil {
		return
	}

	env.Client, err = kubernetes.NewRetryClient(env.Config, client.Options{})
	if err != nil {
		return
	}

	env.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(env.Config)
	return
}

// TestEnvironment is a struct containing the envtest environment, Kubernetes config and clients.
type TestEnvironment struct {
	Environment     *envtest.Environment
	Config          *rest.Config
	Client          Client
	DiscoveryClient discovery.DiscoveryInterface
}

// Client is the controller-runtime Client interface with an added Watch method.
type Client interface {
	client.Client
	// Watch watches a specific object and returns all events for it.
	Watch(ctx context.Context, obj runtime.Object) (watch.Interface, error)
}
