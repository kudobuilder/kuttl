SHELL=/bin/bash -o pipefail

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

.PHONY: prebuild
prebuild: generate lint

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

.PHONY: clean
# Clean all
clean: test-clean

.PHONY: imports
# used to update imports on project.  NOT a linter.
imports:
ifeq (, $(shell which golangci-lint))
	./hack/install-golangcilint.sh
endif
	golangci-lint run --disable-all -E goimports --fix

.PHONY: todo
# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=hack \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow' .
