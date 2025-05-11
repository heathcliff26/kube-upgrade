SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

default: upgraded upgrade-controller

# Build the upgraded binary
upgraded:
	hack/build.sh upgraded

# Build the upgrade controller container image
upgrade-controller:
	podman build -t $(REPOSITORY)/kube-upgrade-controller:$(TAG) -f cmd/upgrade-controller/Dockerfile .

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
lint: golangci-lint
	golangci-lint run -v

# Generate code and artifacts
generate: controller-gen
	hack/generate.sh

# Generate Kubernetes manifests
manifests:
	hack/manifests.sh

# Validate generated files and configurations
validate: controller-gen
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
	rm -rf bin manifests/release coverprofiles coverprofile.out logs tmp_controller_image_kube-upgrade-e2e-*.tar

# Install the golangci-lint tool
golangci-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Install the controller-gen tool
controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.18.0

# Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
	@echo ""
	@echo "Run 'make <target>' to execute a specific target."

.PHONY: \
	default \
	upgraded \
	upgrade-controller \
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
	controller-gen \
	help \
	$(NULL)
