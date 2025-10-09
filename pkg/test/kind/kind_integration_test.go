//go:build integration

package kind

// This test file is the only test that requires docker on the platform.  It may be worth creating a new build tag for that
// but adds complexity for 1 case.

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	dockerClient "github.com/docker/docker/client"
	"github.com/thoas/go-funk"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster/nodes"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

const (
	kindTestContext = "test"
	testImage       = "docker.io/library/busybox:latest"
)

// Tests that Docker images are added to the nodes of a KIND cluster with the
// 'AddContainers' method.
func TestAddContainers(t *testing.T) {
	ctx := context.Background()

	kind := NewKind(kindTestContext, "kubeconfig", testutils.NewTestLogger(t, ""))

	config := v1alpha4.Cluster{}

	if err := kind.Run(&config); err != nil {
		t.Fatalf("failed to start KIND cluster: %v", err)
	}

	t.Cleanup(func() {
		if err := kind.Stop(); err != nil {
			t.Fatalf("failed to stop KIND cluster: %v", err)
		}
	})

	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		t.Fatalf("failed to create Docker client: %v", err)
	}

	docker.NegotiateAPIVersion(ctx)

	if !kind.IsRunning() {
		t.Error("KIND isn't running")
	}

	reader, err := docker.ImagePull(ctx, testImage, image.PullOptions{})
	if err != nil {
		t.Errorf("failed to pull test image: %v", err)
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		t.Log(scanner.Text())
	}

	if err := reader.Close(); err != nil {
		t.Errorf("failed to close image pull output: %v", err)
	}

	if err := kind.AddContainers(docker, []string{testImage}, t); err != nil {
		t.Errorf("failed to add container to KIND cluster: %v", err)
	}

	nodes, err := kind.Provider.ListNodes(kindTestContext)
	if err != nil {
		t.Fatalf("failed to list nodes of KIND cluster: %v", err)
	}

	for _, node := range nodes {
		images, err := nodeImages(node)
		if err != nil {
			t.Errorf("failed to list node images: %v", err)
		}

		if !funk.ContainsString(images, testImage) {
			t.Errorf("failed to find image %s on node %s", testImage, node.String())
		}
	}
}

func nodeImages(node nodes.Node) ([]string, error) {
	var stdout bytes.Buffer

	cmd := node.Command("ctr", "--namespace=k8s.io", "images", "list", "-q")
	cmd.SetStdout(&stdout)

	if err := cmd.Run(); err != nil {
		return []string{}, err
	}

	return strings.Split(stdout.String(), "\n"), nil
}
