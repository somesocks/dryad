# AGENTS.md

## Project overview
- dryad is an experimental package manager/build tool for multi-language package trees.
- This repo is itself a dryad garden; most tests live under `dyd/roots/components/dryad-cli/tests`.

## Key commands
- See all commands: `dryad system commands`
- See help text for a command: `dryad <command> --help`
- Build a root: `dryad root build <path-to-root> --join-stdout --join-stderr --log-level=debug`
- Run a sprout: `dryad sprout run <path-to-sprout> --join-stdout --join-stderr --log-level=debug`

## CLI options
- String options require `--key=value` form. `--key value` errors with “missing value for option --key”.
- `--` switches to passthrough args.

## Roots layout
- Top-level roots: `dyd/roots/components`, `dyd/roots/tools`, `dyd/roots/libraries`, `dyd/roots/releases`.
- `components/dryad-cli` is the main subtree:
  - `source` (CLI source root)
  - `builds/*` (platform builds: native + OS/arch)
  - `tests/*` (CLI test roots)
- `tools/*` are external toolchain roots (go/hugo/caddy/envsubst).
