#!/bin/sh
set -e
set -x

. $(dirname "$0")/binaries-commits

RM_GOPATH=0

TMP_GOPATH=${TMP_GOPATH:-""}

if [ -z "$TMP_GOPATH" ]; then
	export GOPATH="$(mktemp -d)"
	RM_GOPATH=1
else
	export GOPATH="$TMP_GOPATH"
fi

# Upstream crun, without cgroups (dockerd passes --cgroup-manager=disabled), seccomp, systemd, CRIU or eBPF
install_crun() {
	echo "Install crun version $CRUN_COMMIT"
	git clone https://github.com/containers/crun.git "$GOPATH/crun"
	cd "$GOPATH/crun"
	git checkout -q "$CRUN_COMMIT"
	git submodule update --init --recursive
	./autogen.sh
	./configure --disable-systemd --disable-criu --disable-seccomp --disable-dl --disable-bpf
	make LDFLAGS="-all-static"
	strip crun
	cp crun /usr/local/bin/docker-runc
}

install_containerd() {
	echo "Install containerd version $CONTAINERD_COMMIT"
	git clone https://github.com/containerd/containerd.git "$GOPATH/src/github.com/containerd/containerd" -b v0.2.9
	cd "$GOPATH/src/github.com/containerd/containerd"
	git checkout -q "$CONTAINERD_COMMIT"
	make static
	cp bin/containerd /usr/local/bin/docker-containerd
	cp bin/containerd-shim /usr/local/bin/docker-containerd-shim
}

for prog in "$@"
do
	case $prog in
		crun)
			install_crun
			;;

		containerd)
			install_containerd
			;;

		*)
			echo "Usage: $0 [crun|containerd]"
			exit 1

	esac
done

if [ $RM_GOPATH -eq 1 ]; then
	rm -rf "$GOPATH"
fi
