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
GOLANGCI_LINT_VER = "1.45.2"

export GO111MODULE=on

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##############################
# Development                #
##############################

##@ Development

.PHONY: lint
lint: ## Run golangci-lint
ifneq (${GOLANGCI_LINT_VER}, "$(shell ./bin/golangci-lint version --format short 2>&1)")
	@echo "golangci-lint missing or not version '${GOLANGCI_LINT_VER}', downloading..."
	curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/v${GOLANGCI_LINT_VER}/install.sh" | sh -s -- -b ./bin "v${GOLANGCI_LINT_VER}"
endif
	./bin/golangci-lint --timeout 5m run --build-tags integration
	
.PHONY: download
download:  ## Downloads go dependencies
	go mod download

.PHONY: generate-clean
generate-clean:
	rm -rf hack/code-gen

.PHONY: cli
# Build CLI
cli:  ## Builds CLI
	go build -ldflags "${LDFLAGS}" -o bin/${CLI} ./cmd/kubectl-kuttl

.PHONY: cli-clean
# Clean CLI build
cli-clean:
	rm -f bin/${CLI}

.PHONY: clean
clean: cli-clean  ## Cleans CLI and kind logs
	rm -rf kind-logs-*

.PHONY: docker
# build a local docker image (specific to the local platform only)
docker:  ## Builds docker image for architecture of the local env
	docker build . -t kuttl

.PHONY: docker-release
# build and push a multi-arch docker image
docker-release:  ## Build and push multi-arch docker images
	./hack/docker-release.sh


# Install CLI
cli-install:  ## Installs kubectl-kuttl to GOBIN
	go install -ldflags "${LDFLAGS}" ./cmd/kubectl-kuttl

##############################
# Generate Artifacts         #
##############################

##@ Generate

.PHONY: generate
# Generate code
generate: ## Generates code 
ifneq ($(shell go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools), $(shell controller-gen --version 2>/dev/null | cut -b 10-))
	@echo "(Re-)installing controller-gen. Current version:  $(controller-gen --version 2>/dev/null | cut -b 10-). Need $(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)"
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@$$(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$$(go list -f '{{.Version}}' -m sigs.k8s.io/controller-tools)
	go mod tidy
endif
	controller-gen crd paths=./pkg/apis/... output:crd:dir=config/crds output:stdout
	./hack/update_codegen.sh


##############################
# Reports                    #
##############################

##@ Reports

.PHONY: todo
# Show to-do items per file.
todo: ## Shows todos from code
	@grep \
		--exclude-dir=hack \
		--exclude=Makefile \
		--exclude-dir=.git \
		--exclude-dir=bin \
		--text \
		--color \
		-nRo -E " *[^\.]TODO.*|SkipNow" .


##############################
# Tests                      #
##############################

##@ Tests

.PHONY: all
all: lint test integration-test  ## Runs lint, unit and integration tests

# Run unit tests
.PHONY: test
test: ## Runs unit tests
ifdef _INTELLIJ_FORCE_SET_GOFLAGS
# Run tests from a Goland terminal. Goland already set '-mod=readonly'
	go test ./pkg/...  -v -coverprofile cover.out
else
	go test ./pkg/...  -v -mod=readonly -coverprofile cover.out
endif

.PHONY: integration-test
# Run integration tests
integration-test:  ## Runs integration tests
	./hack/run-integration-tests.sh

# Run e2e tests
.PHONY: e2e-test
e2e-test: cli
	./hack/run-e2e-tests.sh
