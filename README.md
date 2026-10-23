# System Docker engine

The daemon BurmillaOS uses to run its system services, forked from Docker
17.06 and stripped to what `ros` needs: containers, images, networks and
volumes over a unix socket, json-file logging and overlay2. There is no
swarm, build, push, plugin, TCP/TLS, seccomp, AppArmor or cgroup support.

Containers run with upstream [crun](https://github.com/containers/crun)
(installed as `system-docker-runc`, cgroups disabled) under containerd 0.2.9.

## Building

```
make binary
```

builds the Debian-based image from `Dockerfile`, which compiles crun and
containerd (pinned in `hack/dockerfile/binaries-commits`), and leaves a
pure-Go static `dockerd` plus `docker-containerd`, `docker-containerd-shim`
and `docker-runc` in `bundles/<version>/binary-daemon/`.
[docker-ce-packaging](https://github.com/burmilla/docker-ce-packaging) turns
that into the `system-docker*` tgz used by BurmillaOS.

Outside the image, with Go 1.27 and `GO111MODULE=off`, the tree must be at
`$GOPATH/src/github.com/docker/docker`; then `hack/make.sh binary` works.
