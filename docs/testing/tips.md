# Tips and Tricks

This document contains some tips and gotchas that can be helpful when writing tests.

## Loading Built Images Into KIND

When KIND clusters are started, you may want to load an image that has not been pushed into the registry. To do this, you can use the `kindContainers` setting on your `TestSuite`.

For example:

```sh
docker build -t myimage .
```

And then in the TestSuite, set:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
startKIND: true
kindContainers:
- myimage
```

When the KIND cluster is launched, the image will be loaded into it.

## Kubernetes Events

Kubernetes events are regular Kubernetes objects and can be asserted on just like any other object:

```yaml
apiVersion: v1
kind: Event
reason: Started
source:
  component: kubelet
involvedObject:
  apiVersion: v1
  kind: Pod
  name: my-pod
```

## Custom Resource Definitions

New Custom Resource Definitions are not immediately available for use in the Kubernetes API until the Kubernetes API has acknowledged them.

If a Custom Resource Definition is being defined inside of a test step, be sure to to wait for the `CustomResourceDefinition` object to appear.

For example, given this Custom Resource Definition in `tests/e2e/crd-test/00-crd.yaml`:

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: mycrds.mycrd.k8s.io
spec:
  group: mycrd.k8s.io
  version: v1alpha1
  names:
    kind: MyCRD
    listKind: MyCRDList
    plural: mycrds
    singular: mycrd
  scope: Namespaced
```

Create the following assert `tests/e2e/crd-test/00-assert.yaml`:

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: mycrds.mycrd.k8s.io
status:
  acceptedNames:
    kind: MyCRD
    listKind: MyCRDList
    plural: mycrds
    singular: mycrd
  storedVersions:
  - v1alpha1
```

And then the CRD can be used in subsequent steps, `tests/e2e/crd-test/01-use.yaml`:

```yaml
apiVersion: mycrd.k8s.io/v1alpha1
kind: MyCRD
spec:
  test: test
```

Note that CRDs created via the `crdDir` test suite configuration are available for use immediately and do not require an assert like this.

## Helm testing

You can test a Helm chart by installing it in either a test step or your test suite:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
commands:
- command: kubectl create serviceaccount -n kube-system tiller
  ignoreFailure: true
- command: kubectl create clusterrolebinding tiller --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
  ignoreFailure: true
- command: helm init --wait --service-account tiller
- command: helm delete --purge memcached
  ignoreFailure: true
- command: helm install --replace --namespace memcached --name nginx stable/memcached
testDirs:
- ./test/integration
startKIND: true
kindNodeCache: true
```

## Image caching in kind

By default, [kind](https://kind.sigs.k8s.io/) does not persist its containerd directory, meaning that on every test run you will have to download all of the images defined in the tests. However, the kuttl test harness supports creating a named Docker volume for each node specified in the kind configuration (or the default node if no nodes or configuration are specified) that will be used for each test run:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
startKIND: true
kindNodeCache: true
testDirs:
- ./test/integration
```

The first time you run the tests, the nodes will download the images, but subsequent runs will used the cached images.

## IDE completion for kuttl configuration files

While there is no currently available K8S controller to handle the kuttl configuration files,
the [kuttl CRD definitions](https://github.com/kudobuilder/kuttl/blob/main/crds/) may be handy for kuttl users to leverage coding assistance for kuttl configuration files in
their favorite IDE.

For intellij IDEA, see [instructions](https://www.jetbrains.com/help/idea/kubernetes.html#crd) for on how to load the CRD files either from:
- a local clone on your desktop
- remote github raw url pointing to the kuttl repository
- from a K8S cluster where you'd register the CRDs (by running `kubectl apply -f <crd_file.yaml>`)

Screenshots in [PR #376](https://github.com/kudobuilder/kuttl/pull/376)