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
test:
	go test -v -race -coverprofile=coverprofile.out.tmp -coverpkg "./pkg/..." ./cmd/... ./pkg/...
	grep -v "zz_generated" "coverprofile.out.tmp" | grep -v "github.com/heathcliff26/kube-upgrade/pkg/client" > "coverprofile.out"
	rm coverprofile.out.tmp

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

# Run end-to-end tests
e2e:
	go test -count=1 -v ./tests/...

# Scan code for vulnerabilities using gosec
gosec:
	gosec -exclude-generated ./...

# Remove build artifacts and temporary files
clean:
	hack/clean.sh

# Install the tools required for building the app
tools:
	GOBIN="$(shell pwd)/bin" go install tool

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
	update-deps \
	update-external-scripts \
	coverprofile \
	lint \
	generate \
	manifests \
	validate \
	fmt \
	gosec \
	clean \
	tools \
	help \
	$(NULL)
