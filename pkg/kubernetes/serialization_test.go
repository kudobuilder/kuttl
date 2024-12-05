package kubernetes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestLoadYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test.yaml")
	assert.Nil(t, err)
	defer tmpfile.Close()

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
	if err != nil {
		t.Fatal(err)
	}

	objs, err := LoadYAMLFromFile(tmpfile.Name())
	assert.Nil(t, err)

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
	actual, err := LoadYAMLFromFile("test_data/prettydiff-actual.yaml")
	assert.NoError(t, err)
	assert.Len(t, actual, 1)
	expected, err := LoadYAMLFromFile("test_data/prettydiff-expected.yaml")
	assert.NoError(t, err)
	assert.Len(t, expected, 1)

	result, err := PrettyDiff(expected[0].(*unstructured.Unstructured), actual[0].(*unstructured.Unstructured))
	assert.NoError(t, err)
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
