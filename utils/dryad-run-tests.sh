#!/usr/bin/env sh

set -eux

dryad sprouts run --scope=none --include=tests --log-level=debug --join-stderr --join-stdout