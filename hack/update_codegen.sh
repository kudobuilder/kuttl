#!/usr/bin/env bash
# Based on https://github.com/kubernetes/code-generator/blob/master/examples/hack/update-codegen.sh

set -o errexit
set -o nounset
set -o pipefail

# The following solution for making code generation work with go modules is
# borrowed and modified from https://github.com/heptio/contour/pull/1010.
# it has been modified to enable caching.
export GO111MODULE=on
VERSION="$(go list -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' -m k8s.io/code-generator | rev | cut -d"-" -f1 | rev)"
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"
CODE_GEN_DIR="${REPO_ROOT}/hack/vendor/code-gen/$VERSION"

# Cleanup cached dirs at old path which will interfere with the recursive grep that gen_helpers does.
rm -rf "${REPO_ROOT}/hack/code-gen/"

if [[ -d ${CODE_GEN_DIR} ]]; then
    echo "Using cached code generator version: $VERSION"
else
    git clone https://github.com/kubernetes/code-generator.git "${CODE_GEN_DIR}"
    git -C "${CODE_GEN_DIR}" reset --hard "${VERSION}"
fi

# Set GOBIN to make gen_helpers install and run binaries in the versioned directory.
export GOBIN="${CODE_GEN_DIR}/bin"

source "${CODE_GEN_DIR}/kube_codegen.sh"
kube::codegen::gen_helpers \
    --boilerplate "${REPO_ROOT}/hack/boilerplate.go.txt" \
    "${REPO_ROOT}"
