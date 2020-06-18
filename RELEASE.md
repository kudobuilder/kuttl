# Development and Release Process

## Development Processes

### Main

The `HEAD` of `main` branch is considered the most active development, to which new features and bug fixes are applied. The expectation is that main will always build and pass tests.
Development is incremental. It is expected that pull-requests are either:
 - complete and fully integrated into code execution path, or
 - have a feature switch which must be enabled until code completion is reached.

`HEAD` on main is expected to be in a state that it could be released.
The next major release will be tagged and released from main.
If the current released version is `v0.3.1` and the next major release is `v0.4.0`, then when it is deemed appropriate `v0.4.0` will be tagged off main, followed by a release.
From that tag work will continue on main.

*If* it is necessary in the future to have a patch release for `v0.4.1`, a branch will be created off the `v0.4.0` tag with the nomenclature of `releases/{major.minor}`, in this example `releases/0.4`.
Branches matching `releases/*` are protected branches in the repository.

The `HEAD` of the release branches follows the same conventions as main.  It is expected that `HEAD` of the release branch is always in a releasable state. The purpose of the release branch is for bug fixes only.  New features should not be targeted to a release branch.

### Documentation

Pull request for documentation (in the `kudo.dev` repo) should be prepared and reviewed at
the same time as the pull request for any user-visible (documentation-worthy) change.
The description for the code change PR should link to the documentation PR (and vice-versa).

The documentation PR should be made against the `next-release` branch, and merged as soon as its
corresponding PR(s) against the `kuttl` repo is/are merged.

### Bug Fixes

Bug fixes are expected to be worked on and applied to `main`.

If the fix is needed for a previously supported release version of KUTTL, then a backport is expected.
The bug fix pull request is expected to be marked with a `backport` label.
If a bug fix has to be backported to a previous version, it is expected when possible that the backport is a `git cherry-pick`, updated if necessary to conform to previous code and architecture.

On occasion a cherry-pick is too much of a burden. In this case, fix the bug as a new pull request against the release branch.
Provide the same title and details of the original pull request for traceability.

## Release Request Process

The KUTTL Project is aiming to do monthly minor releases but the period could be even shorter if necessary. The process is as follows:

1. Every release should have an appointed release manager (RM)
1. RM is responsible for following the process below
1. RM should announce the release in the public [#kudo slack channel](https://kubernetes.slack.com/messages/kudo/) at least two days prior to the date
1. RM makes sure all PRs that need to go into the release are merged prior to the process starting

### Release Process

The official binaries for KUTTL are created using [goreleaser](https://goreleaser.com/). The [.goreleaser.yml](.goreleaser.yml) defines the binaries which are supported for each release. Goreleaser is right now run manually from the RM computer. The following are preconditions for the release process:

1. Ensure you have credential `GITHUB_TOKEN` set.
The env must include `export GITHUB_TOKEN=<personal access token>`.
[Help](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) provided from Github.
The token must grant full access to: `repo`, `write:packages`, `read:packages`.


The release process is:

1. Tag the kuttl repo with expected release `git tag -a v0.2.0 -m "v0.2.0"` and push the tag `git push --tags`.
1. Invoke goreleaser `goreleaser --rm-dist`.
1. Update the GH release with release highlights. There is a draft that contains categorized changes since the last release. It provides categories for highlights, breaking changes, and contributors which should be added the gorelease release notes. The changelog from the draft log is ignored. After the contents are copied, the draft can be deleted.
1. Send an announcement email to [kudobuilder@googlegroups.com](https://groups.google.com/forum/#!forum/kudobuilder) with the subject `[ANNOUNCE] KUTTL $VERSION is released`
1. Run `./hack/generate_krew.sh` and submit the generated `kuttl.yaml` to https://github.com/kubernetes-sigs/krew-index/.


**Note:** If there are issues with the release, any changes to the repository will result in it being considered "dirty" and not in a state to be released.
It is possible outside of the standard release process to build a "snapshot" release using the following command: `goreleaser release --skip-publish --snapshot --rm-dist`
This process will create a `dist` folder with all the build artifacts. The changelog is not created unless a full release is executed. If you are looking to get a "similar" changelog, install [github-release-notes](https://github.com/buchanae/github-release-notes) and execute `github-release-notes -org kudobuilder -repo kuttl -since-latest-release`.


### Cutting a Release Branch

As outlined above, when it is necessary to create a new release branch, it is necessary to update the [circle-ci config](https://github.com/kudobuilder/kuttl/blob/main/.circle-ci/config.yml#L13) to test merges against the correct branch. It is necessary replace all references to `main` with the appropriate release branch.

### Cutting a Patch Release

When cutting a patch release, for example `v0.3.3`, it is necessary to ensure that all bugs fixed on main after `v0.3.2` have landed on the release branch, `releases/0.3` in this case.
