package kubernetes

import (
	"fmt"
	"io"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	appsv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ImpersonateAs holds the username to impersonate in Kubernetes API calls.
var ImpersonateAs = ""

// GetConfig returns a Kubernetes REST configuration with optional user impersonation.
func GetConfig() (*rest.Config, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client config: %w", err)
	}

	if ImpersonateAs != "" {
		cfg.Impersonate.UserName = ImpersonateAs
	}

	return cfg, nil
}

// BuildConfigWithContext creates a Kubernetes REST configuration from a kubeconfig file and context.
func BuildConfigWithContext(kubeconfigPath, context string) (*rest.Config, error) {
	if context == "" {
		// Use default context
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
}

// Kubeconfig converts a rest.Config into a YAML kubeconfig and writes it to w.
// Takes ownership of w and closes it.
func Kubeconfig(cfg *rest.Config, w io.WriteCloser) error {
	var authProvider *appsv1.AuthProviderConfig
	var execConfig *appsv1.ExecConfig
	if cfg.AuthProvider != nil {
		authProvider = &appsv1.AuthProviderConfig{
			Name:   cfg.AuthProvider.Name,
			Config: cfg.AuthProvider.Config,
		}
	}

	if cfg.ExecProvider != nil {
		execConfig = &appsv1.ExecConfig{
			Command:    cfg.ExecProvider.Command,
			Args:       cfg.ExecProvider.Args,
			APIVersion: cfg.ExecProvider.APIVersion,
			Env:        []appsv1.ExecEnvVar{},
		}

		for _, envVar := range cfg.ExecProvider.Env {
			execConfig.Env = append(execConfig.Env, appsv1.ExecEnvVar{
				Name:  envVar.Name,
				Value: envVar.Value,
			})
		}
	}
	err := rest.LoadTLSFiles(cfg)
	if err != nil {
		w.Close() //nolint:errcheck // We haven't written anything, we do not care if this fails, just preventing leaks.
		return err
	}
	proxyURL, err := proxyURLFor(cfg)
	if err != nil {
		w.Close() //nolint:errcheck // We haven't written anything, we do not care if this fails, just preventing leaks.
		return err
	}
	err = json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil).Encode(&appsv1.Config{
		CurrentContext: "cluster",
		Clusters: []appsv1.NamedCluster{
			{
				Name: "cluster",
				Cluster: appsv1.Cluster{
					Server:                   cfg.Host,
					CertificateAuthorityData: cfg.CAData,
					InsecureSkipTLSVerify:    cfg.Insecure,
					ProxyURL:                 proxyURL,
				},
			},
		},
		Contexts: []appsv1.NamedContext{
			{
				Name: "cluster",
				Context: appsv1.Context{
					Cluster:  "cluster",
					AuthInfo: "user",
				},
			},
		},
		AuthInfos: []appsv1.NamedAuthInfo{
			{
				Name: "user",
				AuthInfo: appsv1.AuthInfo{
					ClientCertificateData: cfg.CertData,
					ClientKeyData:         cfg.KeyData,
					Token:                 cfg.BearerToken,
					Username:              cfg.Username,
					Password:              cfg.Password,
					Impersonate:           cfg.Impersonate.UserName,
					ImpersonateGroups:     cfg.Impersonate.Groups,
					ImpersonateUserExtra:  cfg.Impersonate.Extra,
					AuthProvider:          authProvider,
					Exec:                  execConfig,
				},
			},
		},
	}, w)
	if err != nil {
		w.Close() //nolint:errcheck // We do not care if closing failed in addition to writing, just preventing leaks.
		return err
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing newly written kubeconfig file: %w", err)
	}
	return nil
}

// proxyURLFor returns the proxy that cfg uses to reach its API server, or an empty string if there is none.
// rest.Config only keeps the proxy as a function, so it is resolved here to be written back as proxy-url.
func proxyURLFor(cfg *rest.Config) (string, error) {
	if cfg.Proxy == nil {
		return "", nil
	}
	host, _, err := rest.DefaultServerUrlFor(cfg)
	if err != nil {
		return "", fmt.Errorf("error parsing kubernetes server URL: %w", err)
	}
	proxy, err := cfg.Proxy(&http.Request{URL: host})
	if err != nil {
		return "", fmt.Errorf("error resolving kubernetes proxy URL: %w", err)
	}
	if proxy == nil {
		return "", nil
	}
	return proxy.String(), nil
}
