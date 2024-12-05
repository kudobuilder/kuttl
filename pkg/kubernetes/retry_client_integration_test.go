package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testenv TestEnvironment

func TestMain(m *testing.M) {
	var err error

	testenv, err = StartTestEnvironment(false)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()
	testenv.Environment.Stop()
	os.Exit(exitCode)
}

func TestCreateOrUpdate(t *testing.T) {
	// Run the test a bunch of times to try to trigger a conflict and ensure that it handles conflicts properly.
	for i := 0; i < 10; i++ {
		namespaceName := fmt.Sprintf("default-%d", i)
		namespaceObj := NewResource("v1", "Namespace", namespaceName, "default")

		_, err := CreateOrUpdate(context.TODO(), testenv.Client, namespaceObj, true)
		assert.Nil(t, err)

		depToUpdate := WithSpec(t, NewPod("update-me", namespaceName), map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"image": "nginx",
					"name":  "nginx",
				},
			},
		})

		_, err = CreateOrUpdate(context.TODO(), testenv.Client, SetAnnotation(depToUpdate, "test", "hi"), true)
		assert.Nil(t, err)

		quit := make(chan bool)

		go func() {
			for {
				select {
				case <-quit:
					return
				default:
					CreateOrUpdate(context.TODO(), testenv.Client, SetAnnotation(depToUpdate, "test", fmt.Sprintf("%d", i)), false)
					time.Sleep(time.Millisecond * 75)
				}
			}
		}()

		time.Sleep(time.Millisecond * 50)

		_, err = CreateOrUpdate(context.TODO(), testenv.Client, SetAnnotation(depToUpdate, "test", "hello"), true)
		assert.Nil(t, err)

		quit <- true
	}
}

func TestClientWatch(t *testing.T) {
	pod := WithSpec(t, NewPod("my-pod", "default"), map[string]interface{}{
		"containers": []map[string]interface{}{
			{
				"image": "nginx",
				"name":  "nginx",
			},
		},
	})
	gvk := pod.GetObjectKind().GroupVersionKind()

	events, err := testenv.Client.Watch(context.TODO(), pod)
	assert.Nil(t, err)

	go func() {
		assert.Nil(t, testenv.Client.Create(context.TODO(), pod))
		assert.Nil(t, testenv.Client.Update(context.TODO(), pod))
		assert.Nil(t, testenv.Client.Delete(context.TODO(), pod))
	}()

	eventCh := events.ResultChan()

	event := <-eventCh
	assert.Equal(t, watch.EventType("ADDED"), event.Type)
	assert.Equal(t, gvk, event.Object.GetObjectKind().GroupVersionKind())
	assert.Equal(t, client.ObjectKey{Namespace: "default", Name: "my-pod"}, ObjectKey(event.Object))

	event = <-eventCh
	assert.Equal(t, watch.EventType("MODIFIED"), event.Type)
	assert.Equal(t, gvk, event.Object.GetObjectKind().GroupVersionKind())
	assert.Equal(t, client.ObjectKey{Namespace: "default", Name: "my-pod"}, ObjectKey(event.Object))

	event = <-eventCh
	assert.Equal(t, watch.EventType("DELETED"), event.Type)
	assert.Equal(t, gvk, event.Object.GetObjectKind().GroupVersionKind())
	assert.Equal(t, client.ObjectKey{Namespace: "default", Name: "my-pod"}, ObjectKey(event.Object))

	events.Stop()
}
