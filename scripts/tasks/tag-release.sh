#!/usr/bin/env sh

set -eu
# set -x

BASE=$(pwd)

TAG="release-$(cat $BASE/dyd/roots/dryad/src/dyd/traits/version)"

echo "[info] pushing tag $TAG"

git tag $TAG && git push origin $TAG
