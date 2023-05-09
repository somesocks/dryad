#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

dryad garden build --include=dev-shell \
&& dryad stem run $BASE/dyd/sprouts/dev-shell