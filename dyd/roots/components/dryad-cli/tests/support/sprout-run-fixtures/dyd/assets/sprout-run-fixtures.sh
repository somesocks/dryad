#!/usr/bin/env bash
set -euf -o pipefail

SPROUT_FIXTURE_ASSETS=$(dirname "${BASH_SOURCE[0]}")
RUN_TEMP=''
RUN_PID=''
RUN_IDS=''
RUN_ALL_IDS=''

sprout_fixture_fail() {
    printf '[ERROR] %s\n' "$*" >&2
    if [ -n "$RUN_TEMP" ] && [ -f "$RUN_TEMP/stderr" ]; then cat "$RUN_TEMP/stderr" >&2; fi
    return 1
}

sprout_fixture_assert_content() {
    [ -f "$1" ] || sprout_fixture_fail "missing file: $1"
    [ "$(cat "$1")" = "$2" ] || sprout_fixture_fail "unexpected content in $1 (expected $2)"
}

sprout_fixture_assert_absent() {
    [ ! -f "$RUN_GARDEN/results/$1" ] || sprout_fixture_fail "unexpected execution: $1"
}

sprout_fixture_assert_contains() {
    grep -Fq -- "$2" "$1" || sprout_fixture_fail "missing '$2' in $1"
}

sprout_fixture_release() {
    local id
    for id in "$@"; do : > "$RUN_GARDEN/control/release-$id"; done
}

sprout_fixture_cleanup() {
    local status=$? attempt
    if [ "$status" != 0 ] && [ -n "$RUN_TEMP" ] && [ ! -f "$RUN_TEMP/stderr" ] && [ -f "$RUN_TEMP/build.log" ]; then
        cat "$RUN_TEMP/build.log" >&2
    fi
    if [ -n "$RUN_PID" ]; then
        # Release pending as well as running workers before reaping the CLI.
        sprout_fixture_release $RUN_ALL_IDS
        for ((attempt=0; attempt<400; attempt++)); do
            if ! kill -0 "$RUN_PID" 2>/dev/null; then break; fi
            sleep 0.02
        done
        kill "$RUN_PID" 2>/dev/null || true
        wait "$RUN_PID" 2>/dev/null || true
    fi
    if [ -n "$RUN_TEMP" ]; then
        chmod -R u+w "$RUN_TEMP"
        rm -rf "$RUN_TEMP"
    fi
}

sprout_fixture_setup() {
    local name root flavor
    RUN_TEMP=$(mktemp -d 2>/dev/null || mktemp -d -t dryad-sprout-run-test)
    trap sprout_fixture_cleanup EXIT
    RUN_GARDEN="$RUN_TEMP/garden"
    dryad garden create "$RUN_GARDEN" > "$RUN_TEMP/build.log" 2>&1
    cd "$RUN_GARDEN"
    RUN_GARDEN=$(pwd -P)
    for name in "$@"; do
        root="$RUN_GARDEN/dyd/roots/$name"
        dryad root create "$root" --scope=none >> "$RUN_TEMP/build.log" 2>&1
        cp "$SPROUT_FIXTURE_ASSETS/build-sprout-run-worker" "$root/dyd/commands/dyd-root-build"
        cp "$SPROUT_FIXTURE_ASSETS/sprout-run-worker" "$root/dyd/assets/dyd-stem-run"
        chmod +x "$root/dyd/commands/dyd-root-build" "$root/dyd/assets/dyd-stem-run"
        printf '%s' "$name" > "$root/dyd/traits/name"
        mkdir -p "$root/dyd/variants/flavor"
        for flavor in a b c d; do
            printf '%s' true > "$root/dyd/variants/flavor/$flavor"
            RUN_ALL_IDS="$RUN_ALL_IDS $name-$flavor"
        done
        if ! dryad root build "$root" --scope=none --parallel=1 --join-stdout --join-stderr >> "$RUN_TEMP/build.log" 2>&1; then
            return 1
        fi
    done
    sprout_fixture_reset
}

sprout_fixture_reset() {
    rm -rf "$RUN_GARDEN/results" "$RUN_GARDEN/control"
    mkdir -p "$RUN_GARDEN/results" "$RUN_GARDEN/control"
    printf '%s' 0 > "$RUN_GARDEN/control/active"
    printf '%s' 0 > "$RUN_GARDEN/control/peak"
    : > "$RUN_TEMP/stdout"
    : > "$RUN_TEMP/stderr"
}

sprout_fixture_check_status() {
    if [ "$1" = 0 ]; then
        [ "$2" = 0 ] || sprout_fixture_fail "command failed with status $2"
    else
        [ "$2" != 0 ] || sprout_fixture_fail 'expected command to fail'
    fi
}

sprout_fixture_run() {
    local input=$1 expected=$2 status
    shift 2
    sprout_fixture_reset
    printf '%s' "$input" > "$RUN_TEMP/input"
    if dryad "$@" < "$RUN_TEMP/input" > "$RUN_TEMP/stdout" 2> "$RUN_TEMP/stderr"; then
        status=0
    else
        status=$?
    fi
    sprout_fixture_check_status "$expected" "$status"
}

sprout_fixture_start() {
    local input=$1
    RUN_IDS=$2
    shift 2
    sprout_fixture_reset
    printf '%s' "$input" > "$RUN_TEMP/input"
    dryad "$@" < "$RUN_TEMP/input" > "$RUN_TEMP/stdout" 2> "$RUN_TEMP/stderr" &
    RUN_PID=$!
}

sprout_fixture_wait_started() {
    local expected=$1 attempt count id
    for ((attempt=0; attempt<1000; attempt++)); do
        count=0
        for id in $RUN_IDS; do
            if [ -f "$RUN_GARDEN/control/start-$id" ]; then count=$((count + 1)); fi
        done
        if [ "$count" -ge "$expected" ]; then return; fi
        kill -0 "$RUN_PID" 2>/dev/null || sprout_fixture_fail 'CLI exited before workers started'
        sleep 0.02
    done
    sprout_fixture_fail "expected $expected workers to start before release (got $count)"
}

sprout_fixture_finish() {
    local expected=$1 attempt status
    for ((attempt=0; attempt<1000; attempt++)); do
        if ! kill -0 "$RUN_PID" 2>/dev/null; then break; fi
        sleep 0.02
    done
    if kill -0 "$RUN_PID" 2>/dev/null; then sprout_fixture_fail 'CLI did not finish released workers'; fi
    if wait "$RUN_PID"; then status=0; else status=$?; fi
    RUN_PID=''
    sprout_fixture_check_status "$expected" "$status"
}

sprout_fixture_concurrency_case() {
    local expected=$1 input=$2 ids=$3 id
    shift 3
    sprout_fixture_start "$input" "$ids" "$@"
    sprout_fixture_wait_started "$expected"
    sprout_fixture_release $ids
    sprout_fixture_finish 0
    for id in $ids; do sprout_fixture_assert_content "$RUN_GARDEN/results/$id" done; done
    sprout_fixture_assert_content "$RUN_GARDEN/control/peak" "$expected"
    sprout_fixture_assert_content "$RUN_GARDEN/control/active" 0
}

sprout_fixture_wait_failure() {
    local attempt
    for ((attempt=0; attempt<1000; attempt++)); do
        if grep -Fq 'sprout threw error during execution' "$RUN_TEMP/stderr"; then
            kill -0 "$RUN_PID" 2>/dev/null || sprout_fixture_fail 'CLI returned while a started sibling was blocked'
            return
        fi
        kill -0 "$RUN_PID" 2>/dev/null || sprout_fixture_fail 'CLI exited before reporting variant failure'
        sleep 0.02
    done
    sprout_fixture_fail 'variant failure was not reported'
}
