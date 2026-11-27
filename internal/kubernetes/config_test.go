package kubernetes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestKubeconfigKeepsProxyURL(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.yaml")
	require.NoError(t, os.WriteFile(in, []byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    proxy-url: http://proxy.example.org:3128
    server: https://k8s.example.org
  name: development
contexts:
- context:
    cluster: development
    user: developer
  name: development
current-context: development
users:
- name: developer
  user:
    token: secret
`), 0600))

	cfg, err := clientcmd.BuildConfigFromFlags("", in)
	require.NoError(t, err)
	require.NotNil(t, cfg.Proxy, "client-go should load proxy-url into rest.Config")

	out := filepath.Join(dir, "out.yaml")
	f, err := os.Create(out)
	require.NoError(t, err)
	require.NoError(t, Kubeconfig(cfg, f))

	written, err := clientcmd.LoadFromFile(out)
	require.NoError(t, err)
	require.Contains(t, written.Clusters, "cluster")
	assert.Equal(t, "http://proxy.example.org:3128", written.Clusters["cluster"].ProxyURL)
}

func TestKubeconfigWithoutProxy(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yaml")
	f, err := os.Create(out)
	require.NoError(t, err)
	require.NoError(t, Kubeconfig(&rest.Config{Host: "https://k8s.example.org"}, f))

	written, err := clientcmd.LoadFromFile(out)
	require.NoError(t, err)
	require.Contains(t, written.Clusters, "cluster")
	assert.Empty(t, written.Clusters["cluster"].ProxyURL)
}
