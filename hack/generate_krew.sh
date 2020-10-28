#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This script generates a Krew-compatible plugin manifest. It should be run after goreleaser.

VERSION=${VERSION:-$(git describe --tags | sed 's/^v//g')}

# Generate the manifest for a single platform.
function generate_platform {
    ARCH="${2}"
    if [ "${2}" == "amd64" ]; then
        ARCH=x86_64
    elif [ "${2}" == "386" ]; then
        ARCH=i386
    fi

    local sha
    PLATFORM=`uname`
    if [ "$PLATFORM" == 'Darwin' ]; then
      sha=$(curl -L https://github.com/kudobuilder/kuttl/releases/download/v"${VERSION}"/kuttl_"${VERSION}"_"${1}"_"${ARCH}".tar.gz | shasum -a 256 - | awk '{print $1}')
    else
      sha=$(curl -L https://github.com/kudobuilder/kuttl/releases/download/v"${VERSION}"/kuttl_"${VERSION}"_"${1}"_"${ARCH}".tar.gz | sha256sum - | awk '{print $1}')
    fi

    cat <<EOF
  - selector:
      matchLabels:
        os: "${1}"
        arch: "${2}"
    uri: https://github.com/kudobuilder/kuttl/releases/download/v${VERSION}/kuttl_${VERSION}_${1}_${ARCH}.tar.gz
    sha256: "${sha}"
    bin: "${3}"
EOF
}

rm -f kuttl.yaml

# shellcheck disable=SC2129
cat <<EOF >> kuttl.yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: kuttl
spec:
  version: "v${VERSION}"

  shortDescription: Declaratively run and test operators
  homepage: https://kuttl.dev/
  description: |
    The KUbernetes Test TooL (KUTTL) is a highly productive test
    toolkit for testing operators on Kubernetes.
  platforms:
EOF

generate_platform linux amd64 ./kubectl-kuttl >> kuttl.yaml
generate_platform linux 386 ./kubectl-kuttl >> kuttl.yaml
generate_platform darwin amd64 ./kubectl-kuttl >> kuttl.yaml

### KUTTL is not currently built for Windows. Uncomment once it is.
# generate_platform windows amd64 ./kubectl-kuttl.exe >> kuttl.yaml
# generate_platform windows 386 ./kubectl-kuttl.exe >> kuttl.yaml

echo "To publish to the krew index, create a pull request to https://github.com/kubernetes-sigs/krew-index/tree/master/plugins to update kuttl.yaml with the newly generated kuttl.yaml."
