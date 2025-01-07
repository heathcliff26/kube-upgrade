SHELL := bash

REPOSITORY ?= ghcr.io/heathcliff26
TAG ?= latest

default: upgraded upgrade-controller

upgraded:
	hack/build.sh upgraded

upgrade-controller:
	podman build -t $(REPOSITORY)/kube-upgrade-controller:$(TAG) -f cmd/upgrade-controller/Dockerfile .

test:
	go test -v -race -coverprofile=coverprofile.out.tmp -coverpkg "./pkg/..." ./cmd/... ./pkg/...
	grep -v "zz_generated" "coverprofile.out.tmp" | grep -v "github.com/heathcliff26/kube-upgrade/pkg/client" > "coverprofile.out"
	rm coverprofile.out.tmp

update-deps:
	hack/update-deps.sh

update-external-scripts:
	hack/update-external-scripts.sh

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

generate: controller-gen
	hack/generate.sh

manifests:
	hack/manifests.sh

validate: controller-gen
	hack/validate.sh

fmt:
	gofmt -s -w ./cmd ./pkg

e2e:
	go test -count=1 -v ./tests/...

clean:
	rm -rf bin manifests/release coverprofiles coverprofile.out logs tmp_controller_image_kube-upgrade-e2e-*.tar

controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0

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
	clean \
	controller-gen \
	$(NULL)
