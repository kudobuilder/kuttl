#!/usr/bin/env bash

#####
#  Used to produce a multi-arch docker image and push it onto docker hub
#  This requires the use of the buildx experimental feature in the later version of docker
#####

set -o nounset
set -o pipefail
# intentionally not setting 'set -o errexit' because we want to print custom error messages

GIT_VERSION="$(git describe --abbrev=0 --tags | cut -b 2-)"

echo "Releasing for version: $GIT_VERSION"

# run buildx ls to check that buildx is possible on this platform
docker buildx ls > /dev/null 2>&1
RETVAL=$?
if [[ ${RETVAL} != 0 ]]; then
    echo "Invoking 'docker buildx ls' ends with non-zero exit code. (✖╭╮✖)"
    echo "Updated docker with experimental options enabled is required."
    exit 1
fi

docker buildx build . -t "kudobuilder/kuttl:v$GIT_VERSION"  --platform linux/amd64,linux/arm64,linux/ppc64le --push

RETVAL=$?
if [[ ${RETVAL} != 0 ]]; then
    echo "Invoking 'docker buildx build' ends with non-zero exit code. （╯°□°）╯ ┻━┻"
    exit 1
fi

echo "docker build and push was successful! ヽ(•‿•)ノ"
