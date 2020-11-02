#
# This Dockerfile builds a recent cstor-volume-mgmt using the latest binary from
# cstor-volume-mgmt  releases.
#

FROM golang:1.14.7 as build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=on \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT} \
  DEBIAN_FRONTEND=noninteractive \
  PATH="/root/go/bin:${PATH}"

WORKDIR /go/src/github.com/openebs/cstor-operator/

RUN apt-get update && apt-get install -y make git

COPY go.mod go.sum ./
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

COPY . .

RUN make buildx.volume-manager

FROM ubuntu:16.04

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL

LABEL org.label-schema.name="cstor-volume-manager"
LABEL org.label-schema.description="Volume manager for cStor volumes"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

RUN apt-get update; exit 0
RUN apt-get -y install rsyslog

RUN mkdir -p /usr/local/etc/istgt

COPY --from=build /go/src/github.com/openebs/cstor-operator/bin/volume-manager /usr/local/bin/
COPY --from=build /go/src/github.com/openebs/cstor-operator/build/volume-manager/entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT entrypoint.sh
EXPOSE 7676 7777
