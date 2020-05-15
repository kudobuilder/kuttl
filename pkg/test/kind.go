package test

import (
	"context"
	"testing"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha3"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/kind/pkg/log"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// kind provides a thin abstraction layer for a KIND cluster.
type kind struct {
	Provider     *cluster.Provider
	context      string
	explicitPath string
}

type kindLogger struct {
	l testutils.Logger
}

func (k kindLogger) Info(message string) {
	k.l.Log(message)
}

func (k kindLogger) Infof(format string, args ...interface{}) {
	k.l.Logf(format, args...)
}

func (k kindLogger) Enabled() bool {
	return true
}

func (k kindLogger) Warn(message string) {
	k.l.Log(message)
}

func (k kindLogger) Warnf(format string, args ...interface{}) {
	k.l.Logf(format, args...)
}

func (k kindLogger) Error(message string) {
	k.l.Log(message)
}

func (k kindLogger) Errorf(format string, args ...interface{}) {
	k.l.Logf(format, args...)
}

func (k kindLogger) V(level log.Level) log.InfoLogger {
	return k
}

func newKind(kindContext string, explicitPath string, logger testutils.Logger) kind {

	provider := cluster.NewProvider(cluster.ProviderWithLogger(&kindLogger{logger}))

	return kind{
		Provider:     provider,
		context:      kindContext,
		explicitPath: explicitPath,
	}
}

// Run starts a KIND cluster from a given configuration.
func (k *kind) Run(config *v1alpha3.Cluster) error {
	return k.Provider.Create(
		k.context,
		cluster.CreateWithV1Alpha3Config(config),
		cluster.CreateWithKubeconfigPath(k.explicitPath),
		cluster.CreateWithRetain(true),
	)
}

// IsRunning checks if a KIND cluster is already running for the current context.
func (k *kind) IsRunning() bool {
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
func (k *kind) AddContainers(docker testutils.DockerClient, containers []string, t *testing.T) error {
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
func (k *kind) CollectLogs(dir string) error {
	return k.Provider.CollectLogs(k.context, dir)
}

// Stop stops the KIND cluster.
func (k *kind) Stop() error {
	return k.Provider.Delete(k.context, k.explicitPath)
}

func loadContainer(docker testutils.DockerClient, node nodes.Node, container string) error {
	image, err := docker.ImageSave(context.TODO(), []string{container})
	if err != nil {
		return err
	}

	defer image.Close()

	if err := nodeutils.LoadImageArchive(node, image); err != nil {
		return err
	}

	return nil
}
