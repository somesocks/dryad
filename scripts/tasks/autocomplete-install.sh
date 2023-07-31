#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

sudo cp $BASE/dyd/roots/source/bash_autocomplete/dyd/assets/dryad /etc/bash_completion.d/dryad
