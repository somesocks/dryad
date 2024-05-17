#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

dryad run build --scope=docs