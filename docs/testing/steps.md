# Steps

Each test case is broken down into test steps. Test steps within a test case are run sequentially: if any of the test steps fail, the entire test case is considered failed.

A test step can create, update, and delete objects as well as run any kubectl command.

## Format

A test step can include many YAML files and each YAML file can contain many Kubernetes objects. In a test case's directory, each file that begins with the same index is considered a part of the same test step. All objects inside of a test step are operated on by the test harness simultaneously, so use separate test steps to order operations.

E.g., in a test case directory:

```text
tests/e2e/example/00-pod.yaml
tests/e2e/example/00-example.yaml
tests/e2e/example/01-staging.yaml
```

There are two test steps:

* `00`, which includes `00-pod.yaml` and `00-example.yaml`.
* `01`, which includes `01-staging.yaml`.

The test harness would run test step `00` and once completed, run test step `01`.

A namespace is created by the test harness for each test case, so if an object in the step does not have a namespace set, then it will be created in the test case's namespace. If a namespace is set, then that namespace will be respected throughout the tests (making it possible to test resources that reside in standardized namespaces).

See the [configuration reference](reference.md#teststep) for documentation on configuring test steps.

## Creating Objects

Any objects specified in a test step will be created if they do not already exist.

## Updating Objects

If an object does already exist in Kubernetes, then the object in Kubernetes will be updated with the changes specified.

The test harness uses merge patching for updating objects, so it is possible to specify minimal updates. For example, to change the replicas on a deployment but leave all other settings untouched, a step could be written:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 4
```

## Deleting Objects

To delete objects at the beginning of a test step, you can specify object references to delete in your `TestStep` configuration. In a test step file, add a `TestStep` object:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestStep
delete:
# Delete a Pod
- apiVersion: v1
  kind: Pod
  name: my-pod
# Delete all Pods with app=nginx
- apiVersion: v1
  kind: Pod
  labels:
    app: nginx
# Delete all Pods in the test namespace
- apiVersion: v1
  kind: Pod
```

The `delete` object references can delete:

* A single object by specifying its `name`.
* If `labels` is set and `name` is omitted, then objects matching the labels and kind will be deleted.
* If both `name` and `labels` omitted, all objects of the specified kind in the test namespace will be deleted.

The test harness will wait for the objects to be successfully deleted, if they exist, before continuing with the test step - if the objects do not get deleted before the timeout has expired the test step is considered failed.

## Running Commands

A `TestStep` configuration can also specify commands to run before running the step:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestStep
commands:
  - command: kubectl apply -f https://raw.githubusercontent.com/kudobuilder/kudo/master/docs/deployment/10-crds.yaml
    namespaced: true
```

If the `namespaced` setting is set, the `--namespace` flag is set to the test step's namespace.

It is also possible to use any installed kubectl plugin when calling kubectl commands:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestStep
commands:
  - command: kubectl kudo install zookeeper --skip-instance
```

Use defining commands, it is possible to use shell expansion in the command or the scripts the command calls.  Command expansion is the replacement of a variable beginning with `$` with a value from the env such as `$HOME`.  Expands include the OS environment variables.  In addition KUTTL provides or replaces the following:

- `$NAMESPACE` is the namespace kuttl is running the test under
- `$PATH` KUTTL prepends the $PATH with the `$CWD/bin`
- `$KUBECONFIG` is the `$CWD/kubeconfig`

> [!WARNING]
> **Command Expansion of `$`**
>
> The `$` in the command signifies the need for an expansion.
> If you have a need to use `$` without expansion, you will need to escape it by expressing `$$`,
> which will result in a single `$` when the command runs.

### Shell scripts

The command allows only a single binary with parameters to be executed. It does not allow any shell scripting, especially pipes to be used. For bigger problems, it is fine to write a custom shell script and call this via the command, but for simple tasks, KUTTL allows the use of an inline script:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestStep
commands:
  - script: kubectl kudo init --upgrade --dry-run --output yaml | kubectl delete -f -
```

When `script` instead of `command` is used, the attributes `namespaced` is not allowed and silently ignored. You can however use the `$NAMESPACE` environment variable as well as all other ENV vars.

> [!WARNING]
> **Shell dependent behavior**
>
> Scripts are executed by prepending `sh -c` to the given script
> and therefore their behavior depends on the configured environment and shell.
