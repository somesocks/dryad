#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

mkdir -p $HOME/bin
cp $BASE/bootstrap/dryad $HOME/bin/dryad
chmod 755 $HOME/bin/dryad