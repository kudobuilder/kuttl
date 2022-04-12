#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

INTEGRATION_OUTPUT_JUNIT=${INTEGRATION_OUTPUT_JUNIT:-false}
VERSION=${VERSION:-test}

if [ "$INTEGRATION_OUTPUT_JUNIT" == true ]
then
    echo "Running E2E tests with junit output"
    mkdir -p reports/
    go get github.com/jstemmer/go-junit-report
    go install github.com/jstemmer/go-junit-report
    go mod tidy
    
    ./bin/kubectl-kuttl test pkg/test/test_data/ 2>&1 \
        | tee /dev/fd/2 \
        | go-junit-report -set-exit-code \
        > reports/kuttl_e2e_test_report.xml

else
    echo "Running E2E tests without junit output"

    ./bin/kubectl-kuttl test pkg/test/test_data/
fi
