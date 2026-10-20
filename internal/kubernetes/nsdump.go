package kubernetes

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LoggingClient is a Kubernetes client that can also produce log output.
type LoggingClient interface {
	client.Client

	Logf(format string, args ...any)
}

// DumpStuckNamespace logs the state of a namespace that failed to be deleted
// and the objects remaining in it, to make teardown failures diagnosable in CI.
func DumpStuckNamespace(cl LoggingClient, discoveryClient discovery.DiscoveryInterface, nsName string) {
	ns := &corev1.Namespace{}
	if err := cl.Get(context.TODO(), client.ObjectKey{Name: nsName}, ns); err != nil {
		cl.Logf("failed to get namespace %q: %v", nsName, err)
	} else {
		cl.Logf("stuck namespace %q: status=%s deletionTimestamp=%s metadata.finalizers=%v spec.finalizers=%v",
			nsName, ns.Status.Phase, formatDeletionTimestamp(ns.DeletionTimestamp), ns.Finalizers, ns.Spec.Finalizers)
	}

	resources, err := discoveryClient.ServerPreferredNamespacedResources()
	if err != nil {
		cl.Logf("failed to get server preferred namespaced resources: %v", err)
		return
	}

	for _, res := range resources {
		gv, err := schema.ParseGroupVersion(res.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range res.APIResources {
			if !slices.Contains(r.Verbs, "list") {
				continue
			}
			gvk := gv.WithKind(r.Kind)
			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(gvk)

			err = cl.List(context.TODO(), list, client.InNamespace(nsName))
			if err != nil {
				cl.Logf("list %s/%s failed (skip): %v", gv, r.Kind, err)
				continue
			}
			for _, item := range list.Items {
				cl.Logf("remaining %s/%s/%s: deletionTimestamp=%s finalizers=%v",
					gv, r.Kind, item.GetName(), formatDeletionTimestamp(item.GetDeletionTimestamp()), item.GetFinalizers())
			}
		}
	}
}

// formatDeletionTimestamp formats a deletion timestamp for logging:
// quoted because its value contains spaces, "nil" when it is unset.
func formatDeletionTimestamp(ts *metav1.Time) string {
	if ts == nil {
		return "nil"
	}
	return fmt.Sprintf("%q", ts.String())
}
