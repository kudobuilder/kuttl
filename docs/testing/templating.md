# Templating

In most cases, if a test file ends with `.gotmpl.yaml`, then it will be treated as a template for expansion, before it is parsed as YAML.

The data structure available to templates is defined as follows:

```go
type Env struct {
	// Name of the namespace for the test.
	Namespace string
	// Any variables defined when invoking `kuttl test`
	// with --template-var name=value
	Vars map[string]any
}
```

- TODO: example
- TODO: what can the value be