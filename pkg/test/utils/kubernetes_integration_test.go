// +build integration

package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
)

var testenv TestEnvironment

func TestMain(m *testing.M) {
	var err error

	testenv, err = StartTestEnvironment(APIServerDefaultArgs)
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
		depToUpdate := WithSpec(t, NewPod("update-me", fmt.Sprintf("default-%d", i)), map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"image": "nginx",
					"name":  "nginx",
				},
			},
		})

		_, err := CreateOrUpdate(context.TODO(), testenv.Client, SetAnnotation(depToUpdate, "test", "hi"), true)
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
	assert.Equal(t, client.ObjectKey{"default", "my-pod"}, ObjectKey(event.Object))

	event = <-eventCh
	assert.Equal(t, watch.EventType("MODIFIED"), event.Type)
	assert.Equal(t, gvk, event.Object.GetObjectKind().GroupVersionKind())
	assert.Equal(t, client.ObjectKey{"default", "my-pod"}, ObjectKey(event.Object))

	event = <-eventCh
	assert.Equal(t, watch.EventType("DELETED"), event.Type)
	assert.Equal(t, gvk, event.Object.GetObjectKind().GroupVersionKind())
	assert.Equal(t, client.ObjectKey{"default", "my-pod"}, ObjectKey(event.Object))

	events.Stop()
}

func TestRunCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hcmd := harness.Command{
		Command: "echo 'hello'",
	}

	// assert foreground cmd returns nil
	cmd, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.NoError(t, err)
	assert.Nil(t, cmd)
	// foreground processes should have stdout
	assert.NotEmpty(t, stdout)

	hcmd.Background = true
	stdout = &bytes.Buffer{}

	// assert background cmd returns process
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	// no stdout for background processes
	assert.Empty(t, strings.TrimSpace(stdout.String()))

	stdout = &bytes.Buffer{}
	hcmd.Background = false
	hcmd.Command = "sleep 42"

	// assert foreground cmd times out
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 2)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "timeout"))
	assert.Nil(t, cmd)

	stdout = &bytes.Buffer{}
	hcmd.Background = false
	hcmd.Command = "sleep 42"
	hcmd.Timeout = 2

	// assert foreground cmd times out with command timeout
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "timeout"))
	assert.Nil(t, cmd)

}

func TestRunCommandIgnoreErrors(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hcmd := harness.Command{
		Command:       "sleep -u",
		IgnoreFailure: true,
	}

	// assert foreground cmd returns nil
	cmd, err := RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.NoError(t, err)
	assert.Nil(t, cmd)

	hcmd.IgnoreFailure = false
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.Error(t, err)
	assert.Nil(t, cmd)

	// bad commands have errors regardless of ignore setting
	hcmd = harness.Command{
		Command:       "bad-command",
		IgnoreFailure: true,
	}
	cmd, err = RunCommand(context.TODO(), "", hcmd, "", stdout, stderr, 0)
	assert.Error(t, err)
	assert.Nil(t, cmd)
}
