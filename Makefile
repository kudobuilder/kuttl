SHELL=/bin/bash -o pipefail

CLI := kubectl-kuttl
GIT_VERSION_PATH := github.com/kudobuilder/kuttl/pkg/version.gitVersion
GIT_VERSION := $(shell git describe --abbrev=0 --tags | cut -b 2-)
GIT_COMMIT_PATH := github.com/kudobuilder/kuttl/pkg/version.gitCommit
GIT_COMMIT := $(shell git rev-parse HEAD | cut -b -8)
SOURCE_DATE_EPOCH := $(shell git show -s --format=format:%ct HEAD)
BUILD_DATE_PATH := github.com/kudobuilder/kuttl/pkg/version.buildDate
DATE_FMT := "%Y-%m-%dT%H:%M:%SZ"
BUILD_DATE := $(shell date -u -d "@$SOURCE_DATE_EPOCH" "+${DATE_FMT}" 2>/dev/null || date -u -r "${SOURCE_DATE_EPOCH}" "+${DATE_FMT}" 2>/dev/null || date -u "+${DATE_FMT}")
LDFLAGS := -X ${GIT_VERSION_PATH}=${GIT_VERSION} -X ${GIT_COMMIT_PATH}=${GIT_COMMIT} -X ${BUILD_DATE_PATH}=${BUILD_DATE}

export GO111MODULE=on

.PHONY: all
all: test

# Run unit tests
.PHONY: test
test:
ifdef _INTELLIJ_FORCE_SET_GOFLAGS
# Run tests from a Goland terminal. Goland already set '-mod=readonly'
	go test ./pkg/...  -v -coverprofile cover.out
else
	go test ./pkg/...  -v -mod=readonly -coverprofile cover.out
endif

.PHONY: integration-test
# Run integration tests
integration-test:
	./hack/run-integration-tests.sh


.PHONY: lint
lint:
ifeq (, $(shell which golangci-lint))
	./hack/install-golangcilint.sh
endif
	golangci-lint run

.PHONY: download
download:
	go mod download


.PHONY: cli
# Build CLI
cli:
	go build -ldflags "${LDFLAGS}" -o bin/${CLI} ./cmd/kubectl-kuttl

.PHONY: cli-clean
# Clean CLI build
cli-clean:
	rm -f bin/${CLI}

# Install CLI
cli-install:
	go install -ldflags "${LDFLAGS}" ./cmd/kubectl-kuttl

.PHONY: todo
# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=hack \
		--exclude=Makefile \
		--exclude-dir=.git \
		--exclude-dir=bin \
		--text \
		--color \
		-nRo -E " *[^\.]TODO.*|SkipNow" .
