#!/usr/bin/env bash
set -euf -o pipefail

assert_not_empty() {
    argument="$1"
    argument_name="${2:-value}"

    if [ -z "$argument" ]; then
        echo "[ERROR] Fail: $argument_name must not be an empty string." >&2
        return 1
    fi
}
