# Docker Releases

There is strong interest in the community to support a number of architectures with docker images.  One driver is [Operator SDK scorecard](https://sdk.operatorframework.io/docs/advanced-topics/scorecard/scorecard/) which KUTTL provides additional test features.   The following architectures are needed to support scorecard:

* linux/amd64
* linux/arm64
* linux/ppc64le
* linux/s390x

In order to support this, we are using the new and experimental docker [buildx](https://docs.docker.com/engine/reference/commandline/buildx/).  This allows for building and pushing of several architectures to the same docker repository allowing for the client platform to determine preferred architecture for it's pull request. 

In addition to using `buildx` there are two more requirements to make this work:

1. The base image must support all the architectures desired.  For this reason we now use `registry.access.redhat.com/ubi8/ubi-minimal`.
1. The go libraries must support the architectures.

## Building and Pushing Multi-Arch Docker Image

For a detailed understanding, please read the [buildx build documentation](https://docs.docker.com/engine/reference/commandline/buildx_build/).
You will need to `enable` docker CLI experimental features for this to work.

To build and push manually the following is necessary:  `make docker-release`

This will result in a command similar to the following with the version tag replaced with the current version:
`docker buildx build . -t kudobuilder/kuttl:v0.6.1  --platform linux/amd64,linux/arm64,linux/ppc64le --push`

Updates for platform builds are maintained in [hack/docker-release.sh](hack/docker-release.sh).
