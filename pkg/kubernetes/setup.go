package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InstallManifests recurses over ManifestsDir to install all resources defined in YAML manifests.
func InstallManifests(ctx context.Context, c client.Client, dClient discovery.DiscoveryInterface, manifestsDir string, kinds ...runtime.Object) ([]*apiextv1.CustomResourceDefinition, error) {
	crds := []*apiextv1.CustomResourceDefinition{}

	if manifestsDir == "" {
		return crds, nil
	}

	return crds, filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		extensions := map[string]bool{
			".yaml": true,
			".yml":  true,
			".json": true,
		}
		if !extensions[filepath.Ext(path)] {
			return nil
		}

		objs, err := LoadYAMLFromFile(path)
		if err != nil {
			return err
		}

		for _, obj := range objs {
			if len(kinds) > 0 && !MatchesKind(obj, kinds...) {
				var expectedKinds []string
				// it is expected that it is highly unlikely to be here (an unmatched kind)
				// which is the justification for have a loop in a loop
				for _, k := range kinds {
					expectedKinds = append(expectedKinds, k.GetObjectKind().GroupVersionKind().String())
				}
				log.Printf("Skipping resource %s because it does not match expected kinds: %s", obj.GetObjectKind().GroupVersionKind().String(), strings.Join(expectedKinds, ","))
				continue
			}

			objectKey := ObjectKey(obj)
			if objectKey.Namespace == "" {
				if _, _, err := Namespaced(dClient, obj, "default"); err != nil {
					return err
				}
			}

			updated, err := CreateOrUpdate(ctx, c, obj, true)
			if err != nil {
				return fmt.Errorf("error creating resource %s: %w", ResourceID(obj), err)
			}

			action := "created"
			if updated {
				action = "updated"
			}
			// TODO: use test logger instead of Go logger
			log.Println(ResourceID(obj), action)

			newCrd := apiextv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
					APIVersion: obj.GetObjectKind().GroupVersionKind().Version,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.GetName(),
				},
			}
			crds = append(crds, &newCrd)
		}

		return nil
	})
}
