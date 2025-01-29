#!/usr/bin/env sh

set -eu
set -x

# initialize the garden again to add directories not tracked by git
dryad garden create

dryad garden build --scope=none