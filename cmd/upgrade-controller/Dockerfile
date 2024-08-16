###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.22.6@sha256:367bb5295d3103981a86a572651d8297d6973f2ec8b62f716b007860e22cbc25 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH
ARG RELEASE_VERSION

WORKDIR /app

COPY vendor ./vendor
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY pkg ./pkg
COPY hack ./hack

RUN --mount=type=bind,target=/app/.git,source=.git GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh upgrade-controller

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM scratch AS final-stage

WORKDIR /

COPY --from=build-stage /app/bin/upgrade-controller /

USER 1001

ENTRYPOINT ["/upgrade-controller"]

#
# END final-stage
###############################################################################
