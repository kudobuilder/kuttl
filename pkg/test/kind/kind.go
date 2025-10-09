package kind

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// Kind provides a thin abstraction layer for a KIND cluster.
type Kind struct {
	Provider     *cluster.Provider
	context      string
	explicitPath string
}

func NewKind(kindContext string, explicitPath string, logger testutils.Logger) Kind {
	provider := cluster.NewProvider(cluster.ProviderWithLogger(&kindLogger{logger}))

	return Kind{
		Provider:     provider,
		context:      kindContext,
		explicitPath: explicitPath,
	}
}

// Run starts a KIND cluster from a given configuration.
func (k *Kind) Run(config *v1alpha4.Cluster) error {
	return k.Provider.Create(
		k.context,
		cluster.CreateWithV1Alpha4Config(config),
		cluster.CreateWithKubeconfigPath(k.explicitPath),
		cluster.CreateWithRetain(true),
	)
}

// IsRunning checks if a KIND cluster is already running for the current context.
func (k *Kind) IsRunning() bool {
	contexts, err := k.Provider.List()
	if err != nil {
		panic(err)
	}

	for _, context := range contexts {
		if context == k.context {
			return true
		}
	}

	return false
}

// AddContainers loads the named Docker containers into a KIND cluster.
// The cluster must be running for this to work.
func (k *Kind) AddContainers(docker testutils.DockerClient, containers []string, t *testing.T) error {
	if !k.IsRunning() {
		panic("KIND cluster isn't running")
	}

	t.Logf("Adding Containers to KIND...\n")

	nodes, err := k.Provider.ListNodes(k.context)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		for _, container := range containers {
			t.Logf("Add image %s to node %s\n", container, node.String())
			if err := loadContainer(docker, node, container); err != nil {
				return err
			}
		}
	}

	return nil
}

// CollectLogs saves the cluster logs to a directory.
func (k *Kind) CollectLogs(dir string) error {
	return k.Provider.CollectLogs(k.context, dir)
}

// Stop stops the KIND cluster.
func (k *Kind) Stop() error {
	return k.Provider.Delete(k.context, k.explicitPath)
}

func loadContainer(docker testutils.DockerClient, node nodes.Node, container string) error {
	image, err := docker.ImageSave(context.TODO(), []string{container})
	if err != nil {
		return err
	}

	defer image.Close()

	return nodeutils.LoadImageArchive(node, image)
}

// IsMinVersion checks if pass ver is the min required kind version
func IsMinVersion(ver string) bool {
	minVersion := "kind.sigs.k8s.io/v1alpha4"
	comp := version.CompareKubeAwareVersionStrings(minVersion, ver)
	return comp != -1
}
