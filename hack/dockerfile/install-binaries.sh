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

# Target architecture, in Go naming (amd64, arm64, loong64, riscv64); defaults to the build host
GOARCH=${GOARCH:-$(go env GOHOSTARCH)}
export GOARCH
. $(dirname "$0")/cross-env.sh

# Static C libraries crun links against, built for the target into $CROSS_PREFIX
install_json_c() {
	echo "Install json-c version $JSON_C_COMMIT for $GOARCH"
	git clone https://github.com/json-c/json-c.git "$GOPATH/json-c"
	cd "$GOPATH/json-c"
	git checkout -q "$JSON_C_COMMIT"
	cmake -S . -B build \
		-DCMAKE_SYSTEM_NAME=Linux \
		-DCMAKE_SYSTEM_PROCESSOR="${CROSS_TRIPLE%%-*}" \
		-DCMAKE_C_COMPILER="$CC" \
		-DCMAKE_INSTALL_PREFIX="$CROSS_PREFIX" \
		-DCMAKE_INSTALL_LIBDIR=lib \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=OFF \
		-DBUILD_STATIC_LIBS=ON \
		-DBUILD_TESTING=OFF \
		-DBUILD_APPS=OFF \
		-DDISABLE_EXTRA_LIBS=ON
	cmake --build build -j "$(nproc)"
	cmake --install build
}

install_libcap() {
	echo "Install libcap version $LIBCAP_VERSION for $GOARCH"
	mkdir -p "$GOPATH/libcap"
	curl -fsSL "${LIBCAP_URL:-https://mirrors.edge.kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-$LIBCAP_VERSION.tar.xz}" \
		| tar -xJC "$GOPATH/libcap" --strip-components=1
	cd "$GOPATH/libcap"
	make -C libcap CC="$CC" AR="$CROSS_TRIPLE-ar" RANLIB="$CROSS_TRIPLE-ranlib" BUILD_CC=gcc \
		SHARED=no prefix="$CROSS_PREFIX" lib=lib \
		install-static-cap
}

# Upstream crun, without cgroups (dockerd passes --cgroup-manager=disabled), seccomp, systemd, CRIU or eBPF
install_crun() {
	echo "Install crun version $CRUN_COMMIT for $GOARCH"
	git clone https://github.com/containers/crun.git "$GOPATH/crun"
	cd "$GOPATH/crun"
	git checkout -q "$CRUN_COMMIT"
	git submodule update --init --recursive
	./autogen.sh
	PKG_CONFIG_PATH= PKG_CONFIG_LIBDIR="$CROSS_PREFIX/lib/pkgconfig" \
	./configure --host="$CROSS_TRIPLE" CC="$CC" \
		CPPFLAGS="-I$CROSS_PREFIX/include" LDFLAGS="-L$CROSS_PREFIX/lib" \
		--disable-systemd --disable-criu --disable-seccomp --disable-dl --disable-bpf
	make LDFLAGS="-L$CROSS_PREFIX/lib -all-static"
	"$CROSS_TRIPLE-strip" crun
	cp crun /usr/local/bin/docker-runc
}

# Pure Go, so cross compiling only needs GOARCH
install_containerd() {
	echo "Install containerd version $CONTAINERD_COMMIT for $GOARCH"
	git clone https://github.com/burmilla/containerd.git "$GOPATH/src/github.com/containerd/containerd"
	cd "$GOPATH/src/github.com/containerd/containerd"
	git checkout -q "$CONTAINERD_COMMIT"
	CGO_ENABLED=0 make static
	cp bin/containerd /usr/local/bin/docker-containerd
	cp bin/containerd-shim /usr/local/bin/docker-containerd-shim
}

for prog in "$@"
do
	case $prog in
		json-c)
			install_json_c
			;;

		libcap)
			install_libcap
			;;

		crun)
			install_crun
			;;

		containerd)
			install_containerd
			;;

		*)
			echo "Usage: $0 [json-c|libcap|crun|containerd]"
			exit 1

	esac
done

if [ $RM_GOPATH -eq 1 ]; then
	rm -rf "$GOPATH"
fi
