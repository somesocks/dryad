#!/usr/bin/env sh

set -eu

BASE_DIR=$(pwd)

nix-shell $BASE_DIR/shell.nix \
	--run "$@"
