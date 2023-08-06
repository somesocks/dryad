#!/usr/bin/env sh

set -eu
# set -x

SRC_DIR=$DYD_STEM
DEST_DIR=$DYD_BUILD

COMMAND="$1"
FILENAME="$(echo $COMMAND | tr ' ' '-')"
HELP="$(eval "$COMMAND --help")"

COMMAND="$COMMAND" \
HELP="$HELP" \
	envsubst \
	< $SRC_DIR/dyd/assets/page_template.md \
	> $DEST_DIR/dyd/assets/$FILENAME.md
