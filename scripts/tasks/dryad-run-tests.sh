#!/usr/bin/env sh

set -eu
set -x

BASE=$(pwd)

PATH="$BASE/bootstrap/:$PATH"

# run only core tests for now, until the other tests are restructured
dryad sprouts run --scope=none --include=tests/core