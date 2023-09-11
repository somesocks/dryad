#!/usr/bin/env sh

set -e

SYSTEM="$(uname)-$(uname -m)"

case $SYSTEM in
	"Linux-aarch64" | "Linux-arm64")
		RELEASE="https://github.com/somesocks/dryad/releases/download/release-$VERSION/dryad-$VERSION-linux-arm64"
		;;
	"Linux-x86_64" | "Linux-amd64")
		RELEASE="https://github.com/somesocks/dryad/releases/download/release-$VERSION/dryad-$VERSION-linux-amd64"
		;;
	"Darwin-aarch64" | "Darwin-arm64")
		RELEASE="https://github.com/somesocks/dryad/releases/download/release-$VERSION/dryad-$VERSION-darwin-arm64"
		;;
	"Darwin-x86_64" | "Darwin-amd64")
		RELEASE="https://github.com/somesocks/dryad/releases/download/release-$VERSION/dryad-$VERSION-darwin-amd64"
		;;
	*)
		echo "cannot install for system ($SYSTEM)" 1>&2;
		exit 1;
		;;
esac

BIN_DIR="$HOME/bin";
mkdir -p $BIN_DIR;
curl -L -f "$RELEASE" -o "$BIN_DIR/dryad";
chmod 755 "$BIN_DIR/dryad";