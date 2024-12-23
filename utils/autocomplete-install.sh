#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

sudo cp $BASE/dyd/sprouts/components/dryad-bash-autocomplete/dyd/assets/dryad \
    /etc/bash_completion.d/dryad
