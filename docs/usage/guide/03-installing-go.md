---
title: 03 - Installing go
layout: default
nav_order: 2
parent: The Guide
grand_parent: Usage
---

# "Installing" go

We need the go toolchain and compiler in order to be able to compile our server into a binary.  However, we don't want to install go globally, that adds noise to our system, lowers reproducibility, and makes things like freezing the toolchain version more difficult.

Instead, we can create a root to "install" go as a package in our workspace, so that we can use our specific go version for the project.

We can start by adding a new root for the go package, by running `dryad root init ./dyd/roots/tools/go`.

In order to create a go package, we need to download a go release for the system os and architecture.  dryad passes the env vars DYD_OS and DYD_ARCH to stems on execution, which we can use to choose the go release we want to download.  In `./dyd/roots/tools/go/dyd/main`, we can write a script to download the go binaries and extract them into the stem we're building.

```
#!/usr/bin/env bash

#
# turn this on to debug script
# set -x

#
# abort on error
set -euf -o pipefail

SRC_DIR=$DYD_STEM
DEST_DIR="$1"

SYSTEM="$DYD_OS-$DYD_ARCH"

declare -A BUILDS=(
	['darwin-amd64']='https://go.dev/dl/go1.20.darwin-amd64.tar.gz'
	['darwin-arm64']='https://go.dev/dl/go1.20.darwin-arm64.tar.gz'
	['linux-amd64']='https://go.dev/dl/go1.20.linux-amd64.tar.gz'
	['linux-arm64']='https://go.dev/dl/go1.20.linux-arm64.tar.gz'
)

BUILD="${BUILDS[$SYSTEM]}"

echo -n "go" > "$DEST_DIR/dyd/traits/name"
echo -n "$BUILD" > "$DEST_DIR/dyd/traits/source"

# download the binary as the main
TEMP_DIR=$(mktemp -d)

curl -L "$BUILD" -o "$TEMP_DIR/go.tar"

mkdir -p $DEST_DIR/dyd/assets
tar -xf $TEMP_DIR/go.tar -C $DEST_DIR/dyd/assets
rm -rf $TEMP_DIR

cat $SRC_DIR/dyd/assets/main > $DEST_DIR/dyd/main
```

Notice at the end of the build script, we copy a main script to our new package to run.  This is a wrapper script to set up the call to the go binary correctly.  In `./dyd/roots/tools/go/dyd/assets/main`, add:

```
#!/usr/bin/env sh

set -eu
# set -x

PATH=$PATH:$DYD_STEM/dyd/assets/go/bin

# note: the go toolchain does not support being run directly from the dryad heap,
# due to go:embed not allowing symlinks for embedding.
# as a workaround, we need to unpack the go toolchain to a temp directory when we want to use it
# this is an issue with some other build environments as well.
# see:
# https://github.com/golang/go/issues/44507
# https://github.com/bazelbuild/rules_go/issues/3110
# https://github.com/bazelbuild/rules_go/issues/3178
_prepare () {
	OUR_FINGERPRINT=$(cat $DYD_STEM/dyd/fingerprint)

	# use the heap context as a place to safely store the unpacked stem
	CACHE_DIR="$HOME/.cache/dyd-go/$OUR_FINGERPRINT"

	# copy the go toolchain into the cache dir
	if [ ! -d "$CACHE_DIR" ]; then
		mkdir -p "$CACHE_DIR"
		cp -rL $DYD_STEM/dyd/assets/* $CACHE_DIR/
	fi
}

_run () {
	GOROOT=$CACHE_DIR/go \
	$CACHE_DIR/go/bin/go "$@"
}

_exit () {
	# rm -rf $CACHE_DIR || true
	echo -n ""
}

trap _exit INT HUP TERM
_prepare
_run "$@"
_exit

```

With these two scripts in place, `dryad garden build` should correctly download and extract the go toolchain into a stem.  You can verify this by running `dryad sprouts exec --include=go -- version`, which should print out the go version.

