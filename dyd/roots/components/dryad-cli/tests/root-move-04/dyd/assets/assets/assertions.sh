#!/usr/bin/env bash

#
# turn this on to debug script
# set -x

#
# abort on error
# https://sipb.mit.edu/doc/safe-shell/
set -euf -o pipefail

assert_directory_exists() {
    if [ ! -d "$1" ]; then
        echo "[ERROR] Fail: directory '$1' does not exist" 1>&2
        return 1
    fi
    # echo "\[INFO\] Pass: directory '\$1' exists" 1>&2
}

assert_file_exists() {
    if [ ! -f "$1" ]; then
        echo "[ERROR] Fail: file '$1' does not exist or is not a regular file" 1>&2
        return 1
    fi
    # echo "[INFO] Pass: file '$1' exists" 1>&2
}

assert_file_content_equals() {
    file="$1"
    expected_content="$2"

    if [ ! -f "$file" ]; then
        echo "[ERROR] Fail: file '$file' does not exist." 1>&2
        return 1
    fi

    # Read the file content
    actual_content=$(cat "$file")

    if [ "$actual_content" != "$expected_content" ]; then
        echo "[ERROR] Fail: file '$file' content does not match. expected ($expected_content), got ($actual_content)" 1>&2
        return 1
    fi

    # echo "[INFO] Pass: file '$file' content matches the expected content" 1>&2
}

assert_symlink_target() {
    # Assign arguments to variables
    symlink_path=$1
    expected_target=$2

    # Check if the provided path is a symlink
    if [ ! -L "$symlink_path" ]; then
        echo "Error: '$symlink_path' is not a symlink." >&2
        return 1
    fi

    # Get the actual target of the symlink
    actual_target=$(readlink "$symlink_path")

    # Compare the actual target with the expected target
    if [ "$actual_target" != "$expected_target" ]; then
        echo "Assertion failed: expected '$expected_target', but '$symlink_path' points to '$actual_target'." 1>&2
        return 1
    fi

    # Return success if everything matches
    return 0
}
