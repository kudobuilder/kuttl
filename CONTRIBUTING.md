# Contributing Guidelines

First: if you're unsure or afraid of anything, just ask or submit the issue or pull request anyway. You won't be yelled at for giving your best effort. The worst that can happen is that you'll be politely asked to change something. We appreciate any sort of contributions, and don't want a wall of rules to get in the way of that.

However, for those individuals who want a bit more guidance on the best way to contribute to the project, read on. This document will cover what we're looking for. By addressing all the points we're looking for, it raises the chances we can quickly merge or address your contributions.

## Developer Certificate Of Origin

KUTTL requires that contributors sign off on changes submitted to kudobuilder repositories.
The [Developer Certificate of Origin (DCO)](https://developercertificate.org/) is a simple way to certify that you wrote or have the right to submit the code you are contributing to the project.

You sign-off by adding the following to your commit messages:

    This is my commit message

    Signed-off-by: Random J Developer <random@developer.example.org>

Git has a `-s` command line option to do this automatically.

    git commit -s -m 'This is my commit message'

You can find the full text of the DCO here: https://developercertificate.org/

## Contributing Steps

1. Submit an issue describing your proposed change to the repo in question.
2. The [repo owners](https://github.com/kudobuilder/kuttl/blob/main/.github/CODEOWNERS) will respond to your issue promptly.
3. If your proposed change is accepted, and you haven't already done so, sign a Contributor License Agreement (see details above).
4. Fork the desired repo, develop and test your code changes.
5. Submit a pull request.

## How to build KUTTL locally

### Pre-requisites

- Git
- Go `1.18` or later. Note that some [Makefile](Makefile) targets assume that your `$GOBIN` is in your `$PATH`.
- A Kubernetes Cluster running version `1.19` or later (e.g., [kind](https://kind.sigs.k8s.io/) or [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/))
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

#### MacOS (darwin) and Go 1.18

The `controller-gen` used with `make generate` has issues with Go 1.18 on MacOS darwin.  An example output looks like:

```sh
go: downloading golang.org/x/sys v0.0.0-20201112073958-5cba982894dd
# golang.org/x/sys/unix
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/syscall_darwin.1_13.go:29:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.1_13.go:27:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.1_13.go:40:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:28:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:43:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:59:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:75:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:90:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:105:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:121:3: //go:linkname must refer to declared function or variable
../../../../pkg/mod/golang.org/x/sys@v0.0.0-20201112073958-5cba982894dd/unix/zsyscall_darwin_arm64.go:121:3: too many errors
make: *** [generate] Error 2
```
For the build to work you need proper version of `controller-gen` to be installed.  The current work around is:

  1. change the `go.mod` L3 `go 1.18` to `go 1.17`
  1. `make generate`
  1. change `go.mod` back

After the correct controller-gen version is installed.  All future builds will work.  The only time this work-around is necessary is if/when the controller-tools version is updated.  [The fix](https://go-review.googlesource.com/c/sys/+/274573/) has been merged into Go and should be resolved with go 1.18.1.

### Build Instructions

- Get the KUTTL repo: `git clone https://github.com/kudobuilder/kuttl.git`
- `cd kuttl`
- `make cli` to build the CLI

After running this command, CLI will be available in `bin/kubectl-kuttl` and you can invoke the command for example like this `bin/kubectl-kuttl version` (no need to install it as `kubectl` plugin).

### Debugging in VSCode

Click Run > Add Configuration... to open the `launch.json` file. Add below configuration objects to specify the necessary sample tests and assertions in the `args[]`.

```json
{
  // Use IntelliSense to learn about possible attributes.
  // Hover to view descriptions of existing attributes.
  // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
  "version": "0.2.0",
  "configurations": [
    {
      "name": "KUTTL Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/kubectl-kuttl/main.go",
      "args": [
        "test", 
        "sample", 
        "--config=${workspaceFolder}/cmd/kubectl-kuttl/<sample_tests_dir_to_run_with_kuttl>"
      ]
    }
  ]
}
```

### Testing

Use `make all` to run all available tests.
Run just `make` to see what individual targets are available.

The project has settled on the [testify](https://github.com/stretchr/testify) library for testing purposes.
Please use it for new tests as well if possible.

## Community, Discussion, and Support

Learn how to engage with the KUTTL community on the [community page](https://kudo.dev/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/kudo/)
- [Mailing List](https://groups.google.com/d/forum/kudobuilder)

## Code culture

This is a set of practices we try to live by when developing KUTTL. These are just defaults (soft rules). Deviations from them are possible, but have to be justified.

### General guidelines
- Main is always releasable (green CI)
- All feature/bug-fixing work should have an open issue with a description, unless it's something very simple
- Every user-facing feature that is NOT behind a feature gate should have integration or an e2e test

### Pull requests
- One core-team member has to approve the PR to be able to merge (all people listed in `.github/CODEOWNERS` file)
- One approval is enough to merge. However, if there are requests for change they have to be resolved prior to the merge
- Since KUTTL is developed in multiple timezones, try to keep the PR open for everyone to be able to see it (~24h, keep in mind public holidays)
- We prefer squash commits so that all changes from a branch are committed to main as a single commit
- Before you merge, make sure your commit title and description are meaningful. Github by default will list all the individual PR commits when squashing which are rarely insightful. We aim for a clean and meaningful commit history.
- Labels: If your PR includes either **breaking changes** or should get additional attention in the release, add one of these label:
  - `release/highlight` For a big new feature, an important bug fix, the focus of the current release
  - `release/breaking-change`  For anything that breaks backwards compatibility and requires users to take special care when upgrading to the new version
  - `release/bugfix` For noteworthy bugfixes

- For a piece of work that takes >3-5 days, pair with somebody
- When you pair with somebody, don't forget to appreciate their work using [co-authorship](https://help.github.com/en/github/committing-changes-to-your-project/creating-a-commit-with-multiple-authors)
- Open a PR as soon as possible to give everybody a chance to review it
- For PRs that tackle a bigger feature/refactoring schedule a walk-through with the team. PR reviews are a lot more meaningful if reviewers understand your code mental model.

### As a code owner (core team member)
- Schedule a portion of your day to review PRs to appreciate work of others

## Is My Thing an Enhancement?

We are trying to figure out the exact shape of an enhancement. Until then here are a few rough heuristics.

An enhancement is anything that:

- a blog post would be written about after its release (ex. [minikube](https://kubernetes.io/blog/2016/07/minikube-easily-run-kubernetes-locally/), [StatefulSets](https://kubernetes.io/blog/2016/07/thousand-instances-of-cassandra-using-kubernetes-pet-set/), [rkt container runtime](https://kubernetes.io/blog/2016/07/rktnetes-brings-rkt-container-engine-to-kubernetes/))
- requires multiple parties/owners participating to complete
- will be graduating from one stage to another (ex. alpha to beta, beta to GA)
- needs significant effort or changes KUTTL in a significant way (ex. something that would take 10 person-weeks to implement, introduce or redesign a system component)
- impacts the UX or operation of KUTTL substantially such that engineers using KUTTL will need retraining
- users will notice and come to rely on

It is unlikely an enhancement if it is:

- fixing a flaky test
- refactoring code
- performance improvements, which are only visible to users as faster API operations, or faster control loops
- adding error messages or events

If you are not sure, ask someone in the [#kudo](https://kubernetes.slack.com/messages/kudo/) channel on Slack or ping someone listed in [CODEOWNERS](https://github.com/kudobuilder/kudo/blob/main/.github/CODEOWNERS).

### When to Create a New Enhancement Issue

Create an issue in this repo once you:

- have circulated your idea to see if there is interest
   - through Community Meetings, KUTTL/KUDO meetings, KUDO mailing lists, or an issue in github.com/kudobuilder/kuttl
- (optionally) have done a prototype in your own fork
- have identified people who agree to work on the enhancement
  - many enhancements will take several releases to progress through Alpha, Beta, and Stable stages
  - you and your team should be prepared to work on the approx. 9mo - 1 year that it takes to progress to Stable status
- are ready to be the project-manager for the enhancement

## Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](https://github.com/kudobuilder/kudo/blob/main/code-of-conduct.md).
