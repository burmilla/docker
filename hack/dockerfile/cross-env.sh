#!/bin/sh
# Maps the Go target architecture in $GOARCH to the GNU toolchain used for the C parts (crun and its libraries).
# Sourced by the Dockerfile and install-binaries.sh.

case "$GOARCH" in
	amd64)   CROSS_TRIPLE=x86_64-linux-gnu ;;
	arm64)   CROSS_TRIPLE=aarch64-linux-gnu ;;
	loong64) CROSS_TRIPLE=loongarch64-linux-gnu ;;
	riscv64) CROSS_TRIPLE=riscv64-linux-gnu ;;
	*)
		echo "Unsupported GOARCH: $GOARCH (supported: amd64 arm64 loong64 riscv64)" >&2
		exit 1
		;;
esac

CC="$CROSS_TRIPLE-gcc"
CROSS_PREFIX="/usr/local/cross/$CROSS_TRIPLE"
export CROSS_TRIPLE CC CROSS_PREFIX
