SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

default: build

build:
	hack/build.sh upgraded
	hack/build.sh upgrade-controller

image:
	podman build -t $(REPOSITORY)/kube-upgrade-controller:$(TAG) -f cmd/upgrade-controller/Dockerfile .

test:
	go test -v -race ./...

update-deps:
	hack/update-deps.sh

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

generate: controller-gen
	hack/generate.sh

manifests: controller-gen
	hack/manifests.sh

validate: controller-gen
	hack/validate.sh

fmt:
	gofmt -s -w ./cmd ./pkg

clean:
	rm -rf bin manifests/release coverprofiles

controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0

.PHONY: \
	default \
	build \
	image \
	test \
	update-deps \
	coverprofile \
	lint \
	generate \
	manifests \
	validate \
	fmt \
	clean \
	controller-gen \
	$(NULL)
