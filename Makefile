SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

default: build

build:
	hack/build.sh upgrade-operator
	hack/build.sh upgrade-daemon
	hack/build.sh upgrade-handler

images: image-upgrade-operator image-upgrade-daemon image-upgrade-handler

image-upgrade-operator:
	podman build -t $(REPOSITORY)/kube-upgrade-operator:$(TAG) -f cmd/upgrade-operator/Dockerfile .

image-upgrade-daemon:
	podman build -t $(REPOSITORY)/kube-upgrade-daemon:$(TAG) -f cmd/upgrade-daemon/Dockerfile .

image-upgrade-handler:
	podman build -t $(REPOSITORY)/kube-upgrade-handler:$(TAG) -f cmd/upgrade-handler/Dockerfile .

test:
	go test -v -race ./...

update-deps:
	hack/update-deps.sh

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

.PHONY: \
	default \
	build \
	images \
	image-upgrade-operator \
	image-upgrade-daemon \
	image-upgrade-handler \
	test \
	update-deps \
	coverprofile \
	lint \
	$(NULL)
