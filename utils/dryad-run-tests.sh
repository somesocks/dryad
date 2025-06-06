#!/usr/bin/env sh

set -eu
set -x

mkdir -p ./logs

dryad sprouts run \
    --scope=none \
    --include="sprout.path().contains('tests')" \
    --log-level=debug \
    --log-stdout=./logs \
    --log-stderr=./logs
