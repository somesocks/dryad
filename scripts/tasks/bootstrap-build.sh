#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

( \
cd $BASE/dyd/roots/source/dryad/dyd/assets && GO111MODULE=on go build \
	-ldflags "-X=main.Version=0.0.0 -X=main.Fingerprint=0000 -s -w" \
	-o $BASE/bootstrap/dryad \
)