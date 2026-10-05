package harness

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindConfig "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
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

func (d *dockerMock) VolumeCreate(_ context.Context, options client.VolumeCreateOptions) (client.VolumeCreateResult, error) {
	return client.VolumeCreateResult{
		Volume: volume.Volume{
			Mountpoint: fmt.Sprintf("/var/lib/docker/data/%s", options.Name),
		},
	}, nil
}

func (d *dockerMock) ImageSave(context.Context, []string, ...client.ImageSaveOption) (client.ImageSaveResult, error) {
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

func TestExtractRunSelector(t *testing.T) {
	cases := map[string]*struct {
		testCase    *harness.TestCase
		runSelector string
	}{
		"no-test-case-file": {},
		"test-case-file": {
			testCase: &harness.TestCase{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuttl.dev/v1beta1",
					Kind:       "TestCase",
				},
			},
			runSelector: "",
		},
		"test-case-file-selector": {
			testCase: &harness.TestCase{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kuttl.dev/v1beta1",
					Kind:       "TestCase",
				},
				TestRunSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"key": "value",
					},
				},
			},
			runSelector: "key=value",
		},
	}

	entries, err := os.ReadDir("test_data1")
	assert.NoError(t, err)

	for _, entry := range entries {
		tc, ok := cases[entry.Name()]
		assert.Truef(t, ok, "test case not found %q", entry.Name())

		actualTestCase, err := loadTestCaseFile(t, "test_data1", entry)
		assert.NoErrorf(t, err, "load test case file %q", entry.Name())
		assert.Equalf(t, tc.testCase, actualTestCase, "unexpected test case content %q", entry.Name())

		runSelector, err := extractRunSelector(actualTestCase)
		assert.NoErrorf(t, err, "extract run selector %q", entry.Name())
		assert.Equalf(t, tc.runSelector, runSelector.String(), "unexpected run selector %q", entry.Name())
	}
}
