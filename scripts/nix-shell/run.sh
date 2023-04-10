#!/usr/bin/env sh

set -eu

nix-shell scripts/nix-shell/shell.nix \
	--run "$@"
