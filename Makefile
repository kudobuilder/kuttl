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
GOLANGCI_LINT_VER = "1.23.8"

export GO111MODULE=on

.PHONY: all
all: lint test integration-test

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

# Run e2e tests
.PHONY: e2e-test
e2e-test: cli
	./hack/run-e2e-tests.sh

.PHONY: lint
lint:
ifneq (${GOLANGCI_LINT_VER}, "$(shell golangci-lint --version | cut -b 27-32)")
	./hack/install-golangcilint.sh
endif
	golangci-lint run

.PHONY: download
download:
	go mod download

.PHONY: generate
# Generate code
generate:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@$$(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)
endif
	controller-gen crd paths=./pkg/apis/... output:crd:dir=config/crds output:stdout
	./hack/update_codegen.sh

.PHONY: generate-clean
generate-clean:
	rm -rf hack/code-gen

.PHONY: cli
# Build CLI
cli:
	go build -ldflags "${LDFLAGS}" -o bin/${CLI} ./cmd/kubectl-kuttl

.PHONY: cli-clean
# Clean CLI build
cli-clean:
	rm -f bin/${CLI}

.PHONY: clean
clean: cli-clean
	rm -rf kind-logs-*

.PHONY: docker
# build docker image
docker:
	docker build . -t kuttl

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
