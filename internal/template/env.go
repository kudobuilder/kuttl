package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/template"
)

// Env represents the data structure available to test file templates.
type Env struct {
	// Name of the namespace for the test.
	Namespace string
	// Any variables defined when invoking `kuttl test`
	// with --template-var name=value
	Vars map[string]any
	// Please keep docs/testing/templating.md in sync.
}

func (e Env) Clone() (Env, error) {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(e); err != nil {
		return Env{}, err
	}
	clone := Env{}
	if err := json.NewDecoder(&buf).Decode(&clone); err != nil {
		return Env{}, err
	}
	return clone, nil
}

func LoadAndExpand(fileName string, env Env) (io.Reader, error) {
	tpl, err := template.ParseFiles(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text/template file %q: %w", fileName, err)
	}
	buf := &bytes.Buffer{}
	// Clone the env to prevent undocumented communication channel between templates.
	envClone, err := env.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone the templating environment: %w", err)
	}
	if err := tpl.Execute(buf, envClone); err != nil {
		return nil, fmt.Errorf("failed to render text/template from file %q: %w", fileName, err)
	}
	return buf, nil
}
