###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.3 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY . ./

RUN GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh upgraded

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM quay.io/fedora/fedora:42@sha256:096a51c8e4458df9e24725c039baba598b64b33f6a6ca5d586b08463551a93bc AS final-stage

RUN dnf install -y --setopt=install_weak_deps=False \
        rpm-ostree && \
    dnf clean -y all

RUN mkdir /etc/kube-upgraded/ && \
    echo "kubeconfig: /host/etc/kubernetes/kubelet.conf" >> /etc/kube-upgraded/config.yaml && \
    echo "kubeadm: /host/usr/bin/kubeadm" >> /etc/kube-upgraded/config.yaml

WORKDIR /

COPY --from=build-stage /app/bin/upgraded /

ENTRYPOINT ["/upgraded"]

#
# END final-stage
###############################################################################
