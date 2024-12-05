package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRetry(t *testing.T) {
	index := 0

	assert.Nil(t, retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("ignore this error")
		}
		return nil
	}, func(err error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))

	assert.Equal(t, 2, index)
}

func TestRetryWithUnexpectedError(t *testing.T) {
	index := 0

	assert.Equal(t, errors.New("bad error"), retry(context.TODO(), func(context.Context) error {
		index++
		if index == 1 {
			return errors.New("bad error")
		}
		return nil
	}, func(err error) bool { return false }, func(err error) bool {
		return err.Error() == "ignore this error"
	}))
	assert.Equal(t, 1, index)
}

func TestRetryWithNil(t *testing.T) {
	assert.Equal(t, nil, retry(context.TODO(), nil, isJSONSyntaxError))
}

func TestRetryWithNilFromFn(t *testing.T) {
	assert.Equal(t, nil, retry(context.TODO(), func(ctx context.Context) error {
		return nil
	}, isJSONSyntaxError))
}

func TestRetryWithNilInFn(t *testing.T) {
	c := RetryClient{}
	var list client.ObjectList
	assert.Error(t, retry(context.TODO(), func(ctx context.Context) error {
		return c.Client.List(ctx, list)
	}, isJSONSyntaxError))
}

func TestRetryWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	assert.Equal(t, errors.New("error"), retry(ctx, func(context.Context) error {
		return errors.New("error")
	}, func(err error) bool { return true }))
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
