# Templating

In most cases, if a test file ends with `.gotmpl.yaml`, then it will be treated as a [Go text/template](https://pkg.go.dev/text/template) for expansion, before it is parsed as YAML.

The data structure available to templates is defined as follows:

```go
type Env struct {
	// Name of the namespace for the test.
	Namespace string
	// Any variables defined when invoking `kuttl test`
	// with --template-var name=value
	// Note that each value is parsed as YAML, so you can
	// pass the usual data structures as variables.
	Vars map[string]any
}
```

## Example

Let's take the [following test suite](../../test/vars/suite1):

```text
suite1/
└── test1
    ├── 00-apply.gotmpl.yaml
    └── 00-assert.gotmpl.yaml
```

- `00-apply.gotmpl.yaml`:

```gotemplate
apiVersion: v1
kind: Pod
metadata:
  name: foo-{{ .Namespace }}-var1-{{ .Vars.var2 }}-var3
spec:
  containers:
    - name: foo
      image: foo
```
- `00-assert.gotmpl.yaml`:
```gotemplate
apiVersion: v1
kind: Pod
metadata:
  name: foo-{{ .Namespace }}-{{ .Vars.var1 }}-var2-{{ .Vars.var3 }}
```

Then, one can invoke `kuttl test --template-var var1=var1 --template-var var2=var2,var3=var3 suite1` and see the following as expected:

```text
=== CONT  kuttl/harness/test1
    logger.go:42: 12:06:57 | test1 | Creating namespace "kuttl-test-trusted-mackerel"
    logger.go:42: 12:06:57 | test1/0-apply | starting test step 0-apply
    logger.go:42: 12:06:57 | test1/0-apply | Pod:kuttl-test-trusted-mackerel/foo-kuttl-test-trusted-mackerel-var1-var2-var3 created
    logger.go:42: 12:06:57 | test1/0-apply | test step completed 0-apply
    logger.go:42: 12:06:57 | test1 | test1 events from ns kuttl-test-trusted-mackerel:
[...]
--- PASS: kuttl (6.43s)
    --- PASS: kuttl/harness (0.00s)
        --- PASS: kuttl/harness/test1 (0.04s)
PASS
```
