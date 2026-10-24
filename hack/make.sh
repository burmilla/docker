#!/usr/bin/env bash
set -e

# This script builds various binary artifacts from a checkout of the docker
# source code.
#
# Requirements:
# - The current directory should be a checkout of the docker source code
#   (https://github.com/docker/docker). Whatever version is checked out
#   will be built.
# - The VERSION file, at the root of the repository, should exist, and
#   will be used as Docker binary version and package version.
# - The hash of the git commit will also be included in the Docker binary,
#   with the suffix -unsupported if the repository isn't clean.
# - The script is intended to be run inside the docker container specified
#   in the Dockerfile at the root of the source. In other words:
#   DO NOT CALL THIS SCRIPT DIRECTLY.
# - The right way to call this script is to invoke "make" from
#   your checkout of the Docker repository.
#   the Makefile will do a "docker build -t docker ." and then
#   "docker run hack/make.sh" in the resulting image.
#

set -o pipefail

export DOCKER_PKG='github.com/docker/docker'
export SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export MAKEDIR="$SCRIPTDIR/make"
export PKG_CONFIG=${PKG_CONFIG:-pkg-config}

# We're a nice, sexy, little shell script, and people might try to run us;
# but really, they shouldn't. We want to be in a container!
inContainer="AssumeSoInitially"
if [ "$(go env GOHOSTOS)" = 'windows' ]; then
	if [ -z "$FROM_DOCKERFILE" ]; then
		unset inContainer
	fi
else
	if [ "$PWD" != "/go/src/$DOCKER_PKG" ]; then
		unset inContainer
	fi
fi

if [ -z "$inContainer" ]; then
	{
		echo "# WARNING! I don't seem to be running in a Docker container."
		echo "# The result of this command might be an incorrect build, and will not be"
		echo "# officially supported."
		echo "#"
		echo "# Try this instead: make all"
		echo "#"
	} >&2
fi

echo

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
	binary-daemon
)

VERSION=$(< ./VERSION)
! BUILDTIME=$(date --rfc-3339 ns 2> /dev/null | sed -e 's/ /T/')
if [ "$DOCKER_GITCOMMIT" ]; then
	GITCOMMIT="$DOCKER_GITCOMMIT"
elif command -v git &> /dev/null && [ -d .git ] && git rev-parse &> /dev/null; then
	GITCOMMIT=$(git rev-parse --short HEAD)
	if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
		GITCOMMIT="$GITCOMMIT-unsupported"
		echo "#~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
		echo "# GITCOMMIT = $GITCOMMIT"
		echo "# The version you are building is listed as unsupported because"
		echo "# there are some files in the git repository that are in an uncommitted state."
		echo "# Commit these changes, or add to .gitignore to remove the -unsupported from the version."
		echo "# Here is the current list:"
		git status --porcelain --untracked-files=no
		echo "#~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
	fi
else
	echo >&2 'error: .git directory missing and DOCKER_GITCOMMIT not specified'
	echo >&2 '  Please either build with the .git directory accessible, or specify the'
	echo >&2 '  exact (--short) commit hash you are building using DOCKER_GITCOMMIT for'
	echo >&2 '  future accountability in diagnosing build issues.  Thanks!'
	exit 1
fi

if [ "$AUTO_GOPATH" ]; then
	rm -rf .gopath
	mkdir -p .gopath/src/"$(dirname "${DOCKER_PKG}")"
	ln -sf ../../../.. .gopath/src/"${DOCKER_PKG}"
	export GOPATH="${PWD}/.gopath"

fi

if [ ! "$GOPATH" ]; then
	echo >&2 'error: missing GOPATH; please see https://golang.org/doc/code.html#GOPATH'
	echo >&2 '  alternatively, set AUTO_GOPATH=1'
	exit 1
fi

# Use these flags when compiling the tests and final binary

IAMSTATIC='true'
if [ -z "$DOCKER_DEBUG" ]; then
	LDFLAGS='-s -w'
fi

# Pure Go, statically linked: no cgo, so no glibc NSS lookups at runtime
# (osusergo reads /etc/passwd and /etc/group directly, netgo resolves DNS in Go)
export CGO_ENABLED=0
ORIG_BUILDFLAGS=( -tags "autogen netgo osusergo static_build $DOCKER_BUILDTAGS" -trimpath )

BUILDFLAGS=( $BUILDFLAGS "${ORIG_BUILDFLAGS[@]}" )

# a helper to provide ".exe" when it's appropriate
binary_extension() {
	if [ "$(go env GOOS)" = 'windows' ]; then
		echo -n '.exe'
	fi
}

hash_files() {
	while [ $# -gt 0 ]; do
		f="$1"
		shift
		dir="$(dirname "$f")"
		base="$(basename "$f")"
		for hashAlgo in md5 sha256; do
			if command -v "${hashAlgo}sum" &> /dev/null; then
				(
					# subshell and cd so that we get output files like:
					#   $HASH docker-$VERSION
					# instead of:
					#   $HASH /go/src/github.com/.../$VERSION/binary/docker-$VERSION
					cd "$dir"
					"${hashAlgo}sum" "$base" > "$base.$hashAlgo"
				)
			fi
		done
	done
}

bundle() {
	local bundle="$1"; shift
	echo "---> Making bundle: $(basename "$bundle") (in $DEST)"
	source "$SCRIPTDIR/make/$bundle" "$@"
}

copy_binaries() {
	dir="$1"
	# Add nested executables to bundle dir so we have complete set of
	# them available; the Dockerfile builds them for the same GOARCH as dockerd
	if [ -x /usr/local/bin/docker-runc ]; then
		echo "Copying nested executables into $dir"
		for file in containerd containerd-shim runc; do
			cp -f `which "docker-$file"` "$dir/"
			if [ "$2" == "hash" ]; then
				hash_files "$dir/docker-$file"
			fi
		done
	fi
}

install_binary() {
	file="$1"
	target="${DOCKER_MAKE_INSTALL_PREFIX:=/usr/local}/bin/"
	if [ "$(go env GOOS)" == "linux" ]; then
		echo "Installing $(basename $file) to ${target}"
		mkdir -p "$target"
		cp -f -L "$file" "$target"
	else
		echo "Install is only supported on linux"
		return 1
	fi
}

main() {
	# We want this to fail if the bundles already exist and cannot be removed.
	# This is to avoid mixing bundles from different versions of the code.
	mkdir -p bundles
	if [ -e "bundles/$VERSION" ] && [ -z "$KEEPBUNDLE" ]; then
		echo "bundles/$VERSION already exists. Removing."
		rm -fr "bundles/$VERSION" && mkdir "bundles/$VERSION" || exit 1
		echo
	fi

	if [ "$(go env GOHOSTOS)" != 'windows' ]; then
		# Windows and symlinks don't get along well

		rm -f bundles/latest
		ln -s "$VERSION" bundles/latest
	fi

	if [ $# -lt 1 ]; then
		bundles=(${DEFAULT_BUNDLES[@]})
	else
		bundles=($@)
	fi
	for bundle in ${bundles[@]}; do
		export DEST="bundles/$VERSION/$(basename "$bundle")"
		# Cygdrive paths don't play well with go build -o.
		if [[ "$(uname -s)" == CYGWIN* ]]; then
			export DEST="$(cygpath -mw "$DEST")"
		fi
		mkdir -p "$DEST"
		ABS_DEST="$(cd "$DEST" && pwd -P)"
		bundle "$bundle"
		echo
	done
}

main "$@"
