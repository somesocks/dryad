#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

dryad garden build --scope=none --include=docs