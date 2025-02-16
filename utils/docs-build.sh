#!/usr/bin/env sh

set -eu

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

dryad run build --scope=docs