#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

declare -r TRUE=0
declare -r FALSE=1

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

    URL="https://github.com/kyverno/kuttl/releases/download/v${VERSION}/kuttl_${VERSION}_${1}_${ARCH}.tar.gz"
    # check file exists first!
    if ! curl --output /dev/null --silent --head --fail "$URL"; then
      >&2 echo "URL does not exist: $URL"
      return $FALSE
    fi

    local sha
    PLATFORM=$(uname)
    if [ "$PLATFORM" == 'Darwin' ]; then
      sha=$(curl -L "$URL" | shasum -a 256 - | awk '{print $1}')
    else
      sha=$(curl -L "$URL" | sha256sum - | awk '{print $1}')
    fi

    cat <<EOF
  - selector:
      matchLabels:
        os: "${1}"
        arch: "${2}"
    uri: https://github.com/kyverno/kuttl/releases/download/v${VERSION}/kuttl_${VERSION}_${1}_${ARCH}.tar.gz
    sha256: "${sha}"
    bin: "${3}"
EOF
return $TRUE
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
generate_platform linux arm64 ./kubectl-kuttl >> kuttl.yaml
generate_platform darwin amd64 ./kubectl-kuttl >> kuttl.yaml
generate_platform darwin arm64 ./kubectl-kuttl >> kuttl.yaml

### Discontinued support for 32-bit darwin
# generate_platform darwin 386 ./kubectl-kuttl >> kuttl.yaml

### KUTTL is not currently built for Windows. Uncomment once it is.
# generate_platform windows amd64 ./kubectl-kuttl.exe >> kuttl.yaml
# generate_platform windows 386 ./kubectl-kuttl.exe >> kuttl.yaml

echo "Successful!"
echo "To publish to the krew index, create a pull request to https://github.com/kubernetes-sigs/krew-index/tree/master/plugins to update kuttl.yaml with the newly generated kuttl.yaml."
