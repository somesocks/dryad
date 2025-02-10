#!/usr/bin/env sh

set -eu
set -x

dryad sprouts run --scope=none --include=tests --log-level=debug --join-stderr --join-stdout