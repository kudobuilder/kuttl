package http

import (
	"bytes"
	"fmt"
	"io"
	coreHTTP "net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"

	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)

// IsURL returns true if string is an URL
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// ToRuntimeObjects takes a url, pulls the file and returns  []runtime.Object
// url must be a full path to a manifest file.  that file can have multiple runtime objects.
func ToRuntimeObjects(urlPath string) ([]runtime.Object, error) {
	apply := []runtime.Object{}

	buf, err := Read(urlPath)
	if err != nil {
		return nil, err
	}

	objs, err := testutils.LoadYAML(urlPath, buf, "")
	if err != nil {
		return nil, fmt.Errorf("url %q load yaml error", urlPath)
	}
	apply = append(apply, objs...)

	return apply, nil
}

// Read returns a buffer for the file at the url
func Read(urlPath string) (*bytes.Buffer, error) {

	response, err := coreHTTP.Get(urlPath) // nolint:gosec
	if err != nil {
		return nil, err
	}
	if response != nil {
		defer response.Body.Close()
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
