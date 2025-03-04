package test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	kindConfig "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestGetTimeout(t *testing.T) {
	h := Harness{}
	assert.Equal(t, 30, h.GetTimeout())

	h.TestSuite.Timeout = 45
	assert.Equal(t, 45, h.GetTimeout())
}

func TestGetReportName(t *testing.T) {
	h := Harness{}
	assert.Equal(t, "kuttl-report", h.reportName())

	h.TestSuite.ReportName = "special-kuttl-report"
	assert.Equal(t, "special-kuttl-report", h.reportName())
}

type dockerMock struct {
	ImageWriter *io.PipeWriter
	imageReader *io.PipeReader
}

func newDockerMock() *dockerMock {
	reader, writer := io.Pipe()

	return &dockerMock{
		ImageWriter: writer,
		imageReader: reader,
	}
}

func (d *dockerMock) VolumeCreate(_ context.Context, options volume.CreateOptions) (volume.Volume, error) {
	return volume.Volume{
		Mountpoint: fmt.Sprintf("/var/lib/docker/data/%s", options.Name),
	}, nil
}

func (d *dockerMock) NegotiateAPIVersion(_ context.Context) {}

func (d *dockerMock) ImageSave(context.Context, []string, ...client.ImageSaveOption) (io.ReadCloser, error) {
	return d.imageReader, nil
}

func TestAddNodeCaches(t *testing.T) {
	h := Harness{
		T:      t,
		docker: newDockerMock(),
	}

	kindCfg := &kindConfig.Cluster{}
	h.addNodeCaches(h.docker, kindCfg)
	assert.Nil(t, kindCfg.Nodes)

	h.TestSuite.KINDNodeCache = true
	h.addNodeCaches(h.docker, kindCfg)
	assert.NotNil(t, kindCfg.Nodes)
	assert.Equal(t, 1, len(kindCfg.Nodes))
	assert.NotNil(t, kindCfg.Nodes[0].ExtraMounts)
	assert.Equal(t, 1, len(kindCfg.Nodes[0].ExtraMounts))
	assert.Equal(t, "/var/lib/containerd", kindCfg.Nodes[0].ExtraMounts[0].ContainerPath)
	assert.Equal(t, "/var/lib/docker/data/kind-0", kindCfg.Nodes[0].ExtraMounts[0].HostPath)

	kindCfg = &kindConfig.Cluster{
		Nodes: []kindConfig.Node{
			{},
			{},
		},
	}

	h.addNodeCaches(h.docker, kindCfg)
	assert.NotNil(t, kindCfg.Nodes)
	assert.Equal(t, 2, len(kindCfg.Nodes))
	assert.NotNil(t, kindCfg.Nodes[0].ExtraMounts)
	assert.Equal(t, 1, len(kindCfg.Nodes[0].ExtraMounts))
	assert.Equal(t, "/var/lib/containerd", kindCfg.Nodes[0].ExtraMounts[0].ContainerPath)
	assert.Equal(t, "/var/lib/docker/data/kind-0", kindCfg.Nodes[0].ExtraMounts[0].HostPath)
	assert.Equal(t, "/var/lib/docker/data/kind-1", kindCfg.Nodes[1].ExtraMounts[0].HostPath)
}
