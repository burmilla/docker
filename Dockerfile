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

FROM debian:jessie
COPY /sources.list /etc/apt/

# Packaged dependencies
RUN apt-get update && apt-get install -y --force-yes \
	--allow-unauthenticated \
	--no-install-recommends \
	apt-utils \
	automake \
	build-essential \
	cmake \
	curl \
	git \
	iptables \
	libcap-dev \
	libtool \
	net-tools \
	pkg-config \
	tar

# Install seccomp: the version shipped upstream is too old
ENV SECCOMP_VERSION 2.3.2
RUN set -x \
	&& export SECCOMP_PATH="$(mktemp -d)" \
	&& curl --insecure -fsSL "https://github.com/seccomp/libseccomp/releases/download/v${SECCOMP_VERSION}/libseccomp-${SECCOMP_VERSION}.tar.gz" \
		| tar -xzC "$SECCOMP_PATH" --strip-components=1 \
	&& ( \
		cd "$SECCOMP_PATH" \
		&& ./configure --prefix=/usr/local \
		&& make \
		&& make install \
		&& ldconfig \
	) \
	&& rm -rf "$SECCOMP_PATH"

# Install Go
ENV GO_VERSION 1.8.3
RUN curl --insecure -fsSL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
	| tar -xzC /usr/local

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

VOLUME /var/lib/docker
WORKDIR /go/src/github.com/docker/docker
ENV DOCKER_BUILDTAGS seccomp

# Install vndr, runc, containerd, tini, docker-proxy
# Please edit hack/dockerfile/install-binaries.sh to update them.
COPY hack/dockerfile/binaries-commits /tmp/binaries-commits
COPY hack/dockerfile/install-binaries.sh /tmp/install-binaries.sh
RUN /tmp/install-binaries.sh vndr runc containerd tini proxy

# Wrap all commands in the "docker-in-docker" script to allow nested containers
ENTRYPOINT ["hack/dind"]

# Upload docker source
COPY . /go/src/github.com/docker/docker
