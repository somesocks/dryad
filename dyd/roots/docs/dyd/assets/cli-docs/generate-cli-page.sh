#!/usr/bin/env sh

set -eu
# set -x

SRC_PACKAGE=$DYD_STEM
PAGE_TEMPLATE_FILE=$SRC_PACKAGE/dyd/assets/cli-docs/page-template.md
TEMPLATE_FILE=$SRC_PACKAGE/dyd/assets/cli-docs/command-template.md
DEST_FILE=$1

cat "$PAGE_TEMPLATE_FILE" > "$DEST_FILE"

dryad system commands | xargs -I {} sh "$SRC_PACKAGE/dyd/assets/cli-docs/generate-cli-command.sh" \
	"{}" "$TEMPLATE_FILE" "$DEST_FILE"