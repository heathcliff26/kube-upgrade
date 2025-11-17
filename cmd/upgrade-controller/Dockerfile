###############################################################################
# BEGIN build-stage
# Compile the binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.4 AS build-stage

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /app

COPY . ./

RUN GOOS=linux GOARCH="${TARGETARCH}" hack/build.sh upgrade-controller

#
# END build-stage
###############################################################################

###############################################################################
# BEGIN final-stage
# Create final docker image
FROM scratch AS final-stage

LABEL   org.opencontainers.image.authors="heathcliff@heathcliff.eu"

WORKDIR /

COPY --from=build-stage /app/bin/upgrade-controller /

USER 1001

ENTRYPOINT ["/upgrade-controller"]

#
# END final-stage
###############################################################################
