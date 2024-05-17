#!/usr/bin/env bash

#
# turn this on to debug script
# set -x

#
# abort on error
# https://sipb.mit.edu/doc/safe-shell/
set -euf -o pipefail

SCRIPT_START=`date +%s`

_MAKE_BUILD_DIR () {
	#
	# make a temporary working directory
	# this command is linux / osx agnostic
	# https://unix.stackexchange.com/questions/30091/fix-or-alternative-for-mktemp-in-os-x
	echo "[INFO] creating temporary working directory"
	TEMP_DIR=''
	TEMP_DIR=`mktemp -d 2>/dev/null || mktemp -d -t 'build-dir'`
}

_BUILD () {
		rm -rf ./docs
    mkdir -p ./docs
    cp -r ./dyd/sprouts/docs/dyd/assets/site/. ./docs/
    chmod -R 755 ./docs
}

_UNMAKE_BUILD_DIR () {
	#
	# clean up build dir
	echo "[INFO] removing temporary working directory"
	rm -rf $TEMP_DIR
}

_CLEANUP () {
	_UNMAKE_BUILD_DIR || true
	echo ""
}

echo "[INFO] starting build"
trap _CLEANUP ERR EXIT
_MAKE_BUILD_DIR
_BUILD
SCRIPT_END=`date +%s`
SCRIPT_RUNTIME=$((SCRIPT_END-SCRIPT_START))
echo "[INFO] build finished in ${SCRIPT_RUNTIME}s"
echo ""