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

