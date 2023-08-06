#!/usr/bin/env sh

set -eu
# set -x

SRC_DIR=$DYD_STEM
DEST_DIR=$DYD_BUILD


dryad system commands | xargs -d '\n' -I {} sh $SRC_DIR/dyd/assets/generate_cli_page.sh "{}"