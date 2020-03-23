# Setup

To setup the tests locally, you need either:

* Docker

Or:

* Go
* kubebuilder

## Downloading kubebuilder

To setup kubebuilder, fetch the latest release from [Github](https://github.com/kubernetes-sigs/kubebuilder/releases) and extract `etcd` and `kube-apiserver` into `/usr/local/kubebuilder/bin/`.

# Docker only

If you don't want to install kubebuilder and other dependencies of KUTTL locally, you can build KUTTL and run the tests inside a Docker container.

To run tests inside a Docker container, you can just execute:

`./test/run_tests.sh`


# Without Docker

## Running unit tests

Unit tests are written for KUTTL using the standard Go testing library. You can run the unit tests:

```
make test
```

## Running integration tests

Or run all tests:

```
make integration-test
```

## Declarative tests

Most tests written for KUTTL use the [declarative test harness](https://kudo.dev/docs/testing) with the controller-runtime's envtest (which starts `etcd` and `kube-apiserver` locally). This means that tests can be written for and run against KUTTL without requiring a Kubernetes cluster (or even Docker).

### CLI examples

Run all integration tests:

```
go run ./cmd/kubectl-kuttl test
```

Run a specific integration test (e.g., the `patch` test from `test/integration/patch`):

```
go run ./cmd/kubectl-kuttl test --test patch
```

Run tests against a live cluster:

```
go run ./cmd/kubectl-kuttl test --start-control-plane=false
```

Run tests against a live cluster and do not delete resources after running:

```
go run ./cmd/kubectl-kuttl test --start-control-plane=false --skip-delete
```
