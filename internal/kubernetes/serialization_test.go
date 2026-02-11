package kubernetes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestLoadYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "test.yaml")
	require.NoError(t, err)

	err = os.WriteFile(tmpfile.Name(), []byte(`
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: nginx
  name: hello
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
`), 0600)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	objs, err := LoadYAMLFromFile(tmpfile.Name())
	require.NoError(t, err)

	assert.Equal(t, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "nginx",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image": "nginx:1.7.9",
						"name":  "nginx",
					},
				},
			},
		},
	}, objs[0])

	assert.Equal(t, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"name": "hello",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image": "nginx:1.7.9",
						"name":  "nginx",
					},
				},
			},
		},
	}, objs[1])
}

func TestPrettyDiff(t *testing.T) {
	actualObjs, err := LoadYAMLFromFile("test_data/prettydiff-actual.yaml")
	assert.NoError(t, err)
	assert.Len(t, actualObjs, 1)
	expectedObjs, err := LoadYAMLFromFile("test_data/prettydiff-expected.yaml")
	assert.NoError(t, err)
	assert.Len(t, expectedObjs, 1)
	expected, ok := expectedObjs[0].(*unstructured.Unstructured)
	assert.True(t, ok, "expected object should be an *Unstructured")
	actual, ok := actualObjs[0].(*unstructured.Unstructured)
	assert.True(t, ok, "actual object should be an *Unstructured")
	if t.Failed() {
		t.FailNow()
	}

	result, err := PrettyDiff(expected, actual)
	require.NoError(t, err)
	assert.Equal(t, `--- Deployment:/central
+++ Deployment:kuttl-test-thorough-hermit/central
@@ -1,7 +1,35 @@
 apiVersion: apps/v1
 kind: Deployment
 metadata:
+  annotations:
+    email: support@stackrox.com
+    meta.helm.sh/release-name: stackrox-central-services
+    meta.helm.sh/release-namespace: kuttl-test-thorough-hermit
+    owner: stackrox
+  labels:
+    app: central
+    app.kubernetes.io/component: central
+    app.kubernetes.io/instance: stackrox-central-services
+    app.kubernetes.io/managed-by: Helm
+    app.kubernetes.io/name: stackrox
+    app.kubernetes.io/part-of: stackrox-central-services
+    app.kubernetes.io/version: 4.3.x-160-g465d734c11
+    helm.sh/chart: stackrox-central-services-400.3.0-160-g465d734c11
+  managedFields: '[... elided field over 10 lines long ...]'
   name: central
+  namespace: kuttl-test-thorough-hermit
+  ownerReferences:
+  - apiVersion: platform.stackrox.io/v1alpha1
+    blockOwnerDeletion: true
+    controller: true
+    kind: Central
+    name: stackrox-central-services
+    uid: ff834d91-0853-42b3-9460-7ebf1c659f8a
+spec: '[... elided field over 10 lines long ...]'
 status:
-  availableReplicas: 1
+  conditions: '[... elided field over 10 lines long ...]'
+  observedGeneration: 2
+  replicas: 1
+  unavailableReplicas: 1
+  updatedReplicas: 1
 
`, result)
}

func TestMatchesKind(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "test.yaml")
	require.NoError(t, err)

	err = os.WriteFile(tmpfile.Name(), []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: hello
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: hello
`), 0600)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	objs, err := LoadYAMLFromFile(tmpfile.Name())
	require.NoError(t, err)

	crd := NewResource("apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", "", "")
	pod := NewResource("v1", "Pod", "", "")
	svc := NewResource("v1", "Service", "", "")

	assert.False(t, MatchesKind(objs[0], crd))
	assert.True(t, MatchesKind(objs[0], pod))
	assert.True(t, MatchesKind(objs[0], pod, crd))
	assert.True(t, MatchesKind(objs[0], crd, pod))
	assert.False(t, MatchesKind(objs[0], crd, svc))

	assert.True(t, MatchesKind(objs[1], crd))
	assert.False(t, MatchesKind(objs[1], pod))
	assert.True(t, MatchesKind(objs[1], pod, crd))
	assert.True(t, MatchesKind(objs[1], crd, pod))
	assert.False(t, MatchesKind(objs[1], svc, pod))
}
