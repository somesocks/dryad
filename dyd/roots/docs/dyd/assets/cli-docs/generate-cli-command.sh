#!/usr/bin/env sh

set -eu
# set -x

COMMAND=$1
TEMPLATE_FILE=$2
DEST_FILE=$3

HELP="$(eval "$COMMAND --help")"

COMMAND="$COMMAND" \
HELP="$HELP" \
	envsubst \
	< "$2" \
	>> "$3"
