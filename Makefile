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

fmt:
	gofmt -s -w ./cmd ./pkg

clean:
	rm -rf bin

.PHONY: \
	default \
	build \
	image \
	test \
	update-deps \
	coverprofile \
	lint \
	fmt \
	clean \
	$(NULL)
