SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

default: upgraded upgrade-controller

upgraded:
	hack/build.sh upgraded

upgrade-controller:
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
	upgraded \
	upgrade-controller \
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
