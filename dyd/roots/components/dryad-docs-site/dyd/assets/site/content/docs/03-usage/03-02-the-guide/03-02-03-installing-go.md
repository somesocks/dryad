---
title: "03.02.03 - Installing go"
description: "Creating a tooling package."
type: docs
layout: single
---

# "Installing" go

We need the go toolchain and compiler in order to be able to compile our server into a binary.  However, we don't want to install go globally, that adds noise to our system, lowers reproducibility, and makes things like freezing the toolchain version more difficult.

Instead, we can create a root to "install" go as a package in our workspace, so that we can use our specific go version for the project.

We can start by adding a new root for the go package, by running `dryad root init ./dyd/roots/tools/go`.

In order to create a go package, we need to download a go release for the system os and architecture.  dryad passes the env vars DYD_OS and DYD_ARCH to stems on execution, which we can use to choose the go release we want to download.  In `./dyd/roots/tools/go/dyd/commands/default`, we can write a script to download the go binaries and extract them into the stem we're building.

```sh
#!/usr/bin/env bash

#
# turn this on to debug script
# set -x

#
# abort on error
# https://sipb.mit.edu/doc/safe-shell/
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

cat $SRC_DIR/dyd/assets/main > $DEST_DIR/dyd/commands/default

```

Notice at the end of the build script, we copy a main script to our new package to run.  This is a wrapper script to set up the call to the go binary correctly.  In `./dyd/roots/tools/go/dyd/assets/main`, add:

```sh
#!/usr/bin/env sh

set -eu
# set -x

PATH=$DYD_STEM/dyd/assets/go/bin:$PATH

_prepare () {
	echo -n ""
}

_run () {
	GOROOT=$DYD_STEM/dyd/assets/go \
	go "$@"
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

With these two scripts in place, `dryad garden build` should correctly download and extract the go toolchain into a stem.  You can verify this by running `dryad sprouts run --include=go -- version`, which should print out the go version.

