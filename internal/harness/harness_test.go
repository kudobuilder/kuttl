package harness

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	kindConfig "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/kudobuilder/kuttl/internal/kubernetes"
	fakek8s "github.com/kudobuilder/kuttl/internal/kubernetes/fake"
	"github.com/kudobuilder/kuttl/internal/report"
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

type trackingClient struct {
	ctrlclient.Client

	onCreate func()
}

func (c *trackingClient) Create(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
	if c.onCreate != nil {
		c.onCreate()
	}
	return c.Client.Create(ctx, obj, opts...)
}

func newTestSuiteDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		require.NoError(t, os.Mkdir(filepath.Join(dir, name), 0o755))
	}
	return dir
}

func TestRunTestsSequentialWhenParallelIsOne(t *testing.T) {
	suiteDir := newTestSuiteDir(t, "test-1", "test-2")

	var active atomic.Int32
	var maxActive atomic.Int32
	var total atomic.Int32
	overlap := make(chan struct{})
	var closeOverlap sync.Once

	cl := &trackingClient{
		Client: fake.NewClientBuilder().WithScheme(kubernetes.Scheme()).Build(),
		onCreate: func() {
			total.Add(1)
			cur := active.Add(1)
			defer active.Add(-1)

			for {
				prev := maxActive.Load()
				if cur <= prev || maxActive.CompareAndSwap(prev, cur) {
					break
				}
			}

			if cur > 1 {
				closeOverlap.Do(func() { close(overlap) })
				return
			}

			select {
			case <-overlap:
			case <-time.After(50 * time.Millisecond):
			}
		},
	}

	h := Harness{
		TestSuite: harness.TestSuite{
			TestDirs: []string{suiteDir},
			Parallel: 1,
			Suppress: []string{"events"},
		},
		T:       t,
		client:  cl,
		dclient: fakek8s.DiscoveryClient(),
		report:  report.NewSuiteCollection("test"),
	}

	h.RunTests()

	assert.Equal(t, int32(2), total.Load())
	assert.Equal(t, int32(1), maxActive.Load(), "tests should run sequentially when Parallel is 1")
}

func TestRunTestsParallelByDefaultAndWhenGreaterThanOne(t *testing.T) {
	if f := flag.Lookup("test.parallel"); f != nil {
		if n, err := strconv.Atoi(f.Value.String()); err == nil && n < 2 {
			t.Skip("skipping parallel execution test when -parallel < 2")
		}
	}

	for _, tc := range []struct {
		name     string
		parallel int
	}{
		{name: "default (0)", parallel: 0},
		{name: "explicit (>1)", parallel: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			suiteDir := newTestSuiteDir(t, "test-1", "test-2")

			var arrived atomic.Int32
			bothArrived := make(chan struct{})
			var closeBoth sync.Once
			var ranConcurrently atomic.Bool

			cl := &trackingClient{
				Client: fake.NewClientBuilder().WithScheme(kubernetes.Scheme()).Build(),
				onCreate: func() {
					if arrived.Add(1) == 2 {
						ranConcurrently.Store(true)
						closeBoth.Do(func() { close(bothArrived) })
					}
					select {
					case <-bothArrived:
					case <-time.After(5 * time.Second):
					}
				},
			}

			h := Harness{
				TestSuite: harness.TestSuite{
					TestDirs: []string{suiteDir},
					Parallel: tc.parallel,
					Suppress: []string{"events"},
				},
				T:       t,
				client:  cl,
				dclient: fakek8s.DiscoveryClient(),
				report:  report.NewSuiteCollection("test"),
			}

			h.RunTests()

			assert.Equal(t, int32(2), arrived.Load())
			assert.True(t, ranConcurrently.Load(), "tests should run concurrently when Parallel is %d", tc.parallel)
		})
	}
}
