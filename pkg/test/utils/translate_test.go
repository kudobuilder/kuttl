package utils

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func testDecoder(s string) *unstructured.Unstructured {
	b := []byte(s)
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(b), len(b))
	u := &unstructured.Unstructured{}
	if err := d.Decode(u); err != nil {
		fmt.Println(err)
	}
	return u
}

func TestTranslate(t *testing.T) {
	namespace := "foo"
	t.Setenv("NAMESPACE", namespace)
	baz := "bar"
	t.Setenv("BAZ", baz)

	manifestTemplate := `
apiVersion: example.com/v1
kind: CustomResource
metadata:
 name: test
 namespace: $NAMESPACE
status:
 ready: true
spec:
 key1:
  key1: data
  key2: $BAZ
 key2:
  key1: "$NAMESPACE"
`

	manifestTemplated := fmt.Sprintf(`
apiVersion: example.com/v1
kind: CustomResource
metadata:
 name: test
 namespace: %s
status:
 ready: true
spec:
 key1:
  key1: data
  key2: %s
 key2:
  key1: %s
`, namespace, baz, namespace)

	assert.Equal(t, fmt.Sprint(testDecoder(manifestTemplated)), fmt.Sprint(Translate(testDecoder(manifestTemplate))))
}
