# This file describes the standard way to build System Docker, using docker
#
# Usage:
#
# # Assemble the build environment and build the binaries:
# make binary
#
# # Mount your source in an interactive container for quick testing:
# docker run -v `pwd`:/go/src/github.com/docker/docker --privileged -i -t docker bash
#

FROM debian:trixie

# Packaged dependencies
RUN apt-get update && apt-get install -y \
	--no-install-recommends \
	autoconf \
	automake \
	build-essential \
	ca-certificates \
	curl \
	git \
	libcap-dev \
	libjson-c-dev \
	libtool \
	pkg-config \
	python3

# Install Go
ENV GO_VERSION 1.27.3
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-$(dpkg --print-architecture).tar.gz" \
	| tar -xzC /usr/local

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
# The tree is GOPATH/vendor based, not Go modules
ENV GO111MODULE off
ENV GOTOOLCHAIN local

VOLUME /var/lib/docker
WORKDIR /go/src/github.com/docker/docker

# Install crun and containerd
# Please edit hack/dockerfile/install-binaries.sh to update them.
COPY hack/dockerfile/binaries-commits /tmp/binaries-commits
COPY hack/dockerfile/install-binaries.sh /tmp/install-binaries.sh
RUN /tmp/install-binaries.sh crun containerd

# Wrap all commands in the "docker-in-docker" script to allow nested containers
ENTRYPOINT ["hack/dind"]

# Upload docker source
COPY . /go/src/github.com/docker/docker
