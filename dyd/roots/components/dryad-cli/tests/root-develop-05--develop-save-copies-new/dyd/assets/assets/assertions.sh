#!/usr/bin/env bash
set -euf -o pipefail

assert_file_exists() {
    if [ ! -f "$1" ]; then
        echo "[ERROR] Fail: file '$1' does not exist or is not a regular file" 1>&2
        return 1
    fi
}

assert_file_content_equals() {
    file="$1"
    expected_content="$2"

    if [ ! -f "$file" ]; then
        echo "[ERROR] Fail: file '$file' does not exist." 1>&2
        return 1
    fi

    actual_content=$(cat "$file")

    if [ "$actual_content" != "$expected_content" ]; then
        echo "[ERROR] Fail: file '$file' content does not match. expected ($expected_content), got ($actual_content)" 1>&2
        return 1
    fi
}
