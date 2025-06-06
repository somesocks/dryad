#!/usr/bin/env sh

set -eu
set -x

mkdir -p ./logs/build

dryad roots build \
    --scope=none \
    --log-level=debug \
    --log-stdout=./logs/build \
    --log-stderr=./logs/build
