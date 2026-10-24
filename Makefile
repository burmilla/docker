.PHONY: binary build help install shell

DOCKER_GITCOMMIT := $(shell git rev-parse --short HEAD || echo unsupported)
export DOCKER_GITCOMMIT

# target architecture in Go naming (amd64, arm64, loong64, riscv64), defaults to the build host
# e.g. `make ARCH=riscv64 binary` cross compiles on any host
ARCH ?= $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/' -e 's/loongarch64/loong64/')

# env vars passed through directly to the build scripts
# to allow things like `make KEEPBUNDLE=1 binary` easily
DOCKER_ENVS := \
	-e BUILDFLAGS \
	-e KEEPBUNDLE \
	-e DOCKER_BUILD_GOGC \
	-e DOCKER_DEBUG \
	-e DOCKER_GITCOMMIT \
	-e HTTP_PROXY \
	-e HTTPS_PROXY \
	-e NO_PROXY \
	-e http_proxy \
	-e https_proxy \
	-e no_proxy

# to allow `make BIND_DIR=. shell`
# (default to no bind mount if DOCKER_HOST is set)
BIND_DIR := $(if $(BINDDIR),$(BINDDIR),$(if $(DOCKER_HOST),,bundles))
DOCKER_MOUNT := $(if $(BIND_DIR),-v "$(CURDIR)/$(BIND_DIR):/go/src/github.com/docker/docker/$(BIND_DIR)")
DOCKER_MOUNT := $(if $(DOCKER_MOUNT),$(DOCKER_MOUNT),-v /go/src/github.com/docker/docker/bundles) -v $(CURDIR)/.git:/go/src/github.com/docker/docker/.git

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
DOCKER_IMAGE := docker-dev:$(if $(GIT_BRANCH_CLEAN),$(GIT_BRANCH_CLEAN)-)$(ARCH)

DOCKER_FLAGS := docker run --rm -i $(DOCKER_ENVS) $(DOCKER_MOUNT)

# if this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

DOCKER_RUN_DOCKER := $(DOCKER_FLAGS) "$(DOCKER_IMAGE)"

default: binary

binary: build ## build the linux binaries
	$(DOCKER_RUN_DOCKER) hack/make.sh binary

build: bundles
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARCH=$(ARCH) -t "$(DOCKER_IMAGE)" .

bundles:
	mkdir bundles

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## install the linux binaries
	KEEPBUNDLE=1 hack/make.sh install-binary

shell: build ## start a shell inside the build env
	$(DOCKER_RUN_DOCKER) bash
