###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.4 AS build-stage

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
FROM quay.io/fedora/fedora:43 AS final-stage

# Force systemd to connect via the system dbus socket, instead of PID 1 socket
ENV SYSTEMCTL_FORCE_BUS=1

RUN dnf install -y --setopt=install_weak_deps=False \
        rpm-ostree \
        hostname && \
    dnf clean -y all

WORKDIR /

COPY --from=build-stage /app/bin/upgraded /app/cmd/upgraded/start-upgraded.sh /

ENTRYPOINT ["/start-upgraded.sh"]

#
# END final-stage
###############################################################################
