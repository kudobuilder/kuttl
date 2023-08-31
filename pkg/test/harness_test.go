package test

import (
	"context"
	"fmt"
	"io"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	kindConfig "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/report"
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

func TestHarnessReport(t *testing.T) {
	type HarnessTest struct {
		name           string
		expectedFormat string
		h              *Harness
	}

	tests := []HarnessTest{
		{
			name:           "should create an XML report when format is XML",
			expectedFormat: "xml",
			h: &Harness{
				TestSuite: harness.TestSuite{
					ReportFormat: "XML",
				},
				report: &report.Testsuites{},
			},
		}, {
			name:           "should create an XML report when format is xml",
			expectedFormat: "xml",
			h: &Harness{
				TestSuite: harness.TestSuite{
					ReportFormat: "xml",
				},
				report: &report.Testsuites{},
			},
		}, {
			name:           "should create an JSON report when format is JSON",
			expectedFormat: "json",
			h: &Harness{
				TestSuite: harness.TestSuite{
					ReportFormat: "JSON",
				},
				report: &report.Testsuites{},
			},
		}, {
			name:           "should create an JSON report when format is json",
			expectedFormat: "json",
			h: &Harness{
				TestSuite: harness.TestSuite{
					ReportFormat: "json",
				},
				report: &report.Testsuites{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the artifacts dir for current test run
			tt.h.TestSuite.ArtifactsDir = t.TempDir()
			tt.h.Report()
			assert.FileExists(t, fmt.Sprintf("%s/%s.%s", tt.h.TestSuite.ArtifactsDir, "kuttl-report", tt.expectedFormat))
		})
	}

	// unit test for not passing any report format
	emptyTest := HarnessTest{
		name:           "should not create any report when format is empty",
		expectedFormat: "json",
		h: &Harness{
			TestSuite: harness.TestSuite{},
			report:    &report.Testsuites{},
		},
	}
	t.Run(emptyTest.name, func(t *testing.T) {
		emptyTest.h.TestSuite.ArtifactsDir = t.TempDir()
		emptyTest.h.Report()
		assert.NoFileExists(t, fmt.Sprintf("%s/%s.%s", emptyTest.h.TestSuite.ArtifactsDir, "kuttl-report", emptyTest.expectedFormat))
	})
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

func (d *dockerMock) VolumeCreate(_ context.Context, body volumetypes.VolumeCreateBody) (dockertypes.Volume, error) {
	return dockertypes.Volume{
		Mountpoint: fmt.Sprintf("/var/lib/docker/data/%s", body.Name),
	}, nil
}

func (d *dockerMock) NegotiateAPIVersion(_ context.Context) {}

func (d *dockerMock) ImageSave(_ context.Context, _ []string) (io.ReadCloser, error) {
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
