# This file describes the standard way to build System Docker, using docker
#
# Usage:
#
# # Assemble the build environment and build the binaries:
# make binary
#
# # Mount your source in an interactive container for quick testing:
# docker run -v `pwd`:/go/src/github.com/docker/docker -i -t docker bash
#

FROM debian:trixie

# Target architecture, in Go naming: amd64, arm64, loong64 or riscv64.
# The build environment always runs natively and cross compiles for ARCH; empty means the build host.
ARG ARCH

# Packaged dependencies
RUN apt-get update && apt-get install -y \
	--no-install-recommends \
	autoconf \
	automake \
	build-essential \
	ca-certificates \
	cmake \
	curl \
	git \
	libtool \
	pkg-config \
	python3 \
	xz-utils

# C cross compiler for crun and its static libraries
COPY hack/dockerfile/cross-env.sh /tmp/cross-env.sh
RUN GOARCH="${ARCH:-$(dpkg --print-architecture)}" && . /tmp/cross-env.sh \
	&& if ! command -v "$CC" > /dev/null; then \
		case "$CROSS_TRIPLE" in \
			loongarch64-*) \
				apt-get install -y --no-install-recommends "gcc-14-$CROSS_TRIPLE" \
				&& ln -s "$CROSS_TRIPLE-gcc-14" "/usr/bin/$CC" ;; \
			*) apt-get install -y --no-install-recommends "gcc-$CROSS_TRIPLE" ;; \
		esac \
	fi

# Install Go
ENV GO_VERSION 1.27.3
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-$(dpkg --print-architecture).tar.gz" \
	| tar -xzC /usr/local

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
# The tree is GOPATH/vendor based, not Go modules
ENV GO111MODULE off
ENV GOTOOLCHAIN local
# Everything Go below (containerd, dockerd) is built for the target
ENV GOARCH=${ARCH}

WORKDIR /go/src/github.com/docker/docker

# Install crun (with static json-c and libcap) and containerd for the target
# Please edit hack/dockerfile/binaries-commits to update them.
COPY hack/dockerfile/binaries-commits /tmp/binaries-commits
COPY hack/dockerfile/install-binaries.sh /tmp/install-binaries.sh
RUN /tmp/install-binaries.sh json-c libcap crun containerd

# Upload docker source
COPY . /go/src/github.com/docker/docker
