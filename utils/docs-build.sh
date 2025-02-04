#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

# initialize the garden again to add directories not tracked by git
dryad garden create

dryad run build --scope=docs