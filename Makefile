SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

# Build all images
build: build-upgraded build-upgrade-controller

# Build the upgraded container image
build-upgraded:
	podman build -t $(REPOSITORY)/kube-upgraded:$(TAG) -f cmd/upgraded/Dockerfile .

# Build the upgrade controller container image
build-upgrade-controller:
	podman build -t $(REPOSITORY)/kube-upgrade-controller:$(TAG) -f cmd/upgrade-controller/Dockerfile .

# Build and push all images
push: push-upgraded push-upgrade-controller

# Build and push upgraded container image
push-upgraded: build-upgraded
	podman push $(REPOSITORY)/kube-upgraded:$(TAG)

# Build and push upgrade controller container image
push-upgrade-controller: build-upgrade-controller
	podman push $(REPOSITORY)/kube-upgrade-controller:$(TAG)

# Run unit-tests
test: tools
	hack/unit-test.sh

# Run end-to-end tests
test-e2e:
	go test -count=1 -v ./tests/...

# Update project dependencies
update-deps:
	hack/update-deps.sh

# Update external scripts used in the project
update-external-scripts:
	hack/update-external-scripts.sh

# Generate coverage profile
coverprofile:
	hack/coverprofile.sh

# Run linter
lint:
	golangci-lint run -v

# Lint the helm charts
lint-helm:
	helm lint manifests/helm/

# Generate code and artifacts
generate: tools
	hack/generate.sh

# Generate Kubernetes manifests
manifests:
	hack/manifests.sh

# Validate generated files and configurations
validate:
	hack/validate.sh

# Format the codebase
fmt:
	gofmt -s -w ./cmd ./pkg

# Scan code for vulnerabilities using gosec
gosec:
	gosec -exclude-generated ./...

# Remove build artifacts and temporary files
clean:
	hack/clean.sh

# Install the tools required for building the app
tools:
	GOBIN="$(shell pwd)/bin" go install tool
	GOBIN="$(shell pwd)/bin" go install github.com/sigstore/cosign/v3/cmd/cosign@latest

# Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

.PHONY: \
	build \
	build-upgraded \
	build-upgrade-controller \
	push \
	push-upgraded \
	push-upgrade-controller \
	test \
	test-e2e \
	update-deps \
	update-external-scripts \
	coverprofile \
	lint \
	lint-helm \
	generate \
	manifests \
	validate \
	fmt \
	gosec \
	clean \
	tools \
	help \
	$(NULL)
