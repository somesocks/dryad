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

## Dryad concepts - Roots
- A root is a build source that produces a stem; roots live under `dyd/roots/...`.
- Root type marker: `dyd/type` with content `root` (no trailing newline).
- Build entrypoint: `dyd/commands/dyd-root-build`.
- Inputs and metadata: `dyd/assets/`, `dyd/docs/`, `dyd/traits/`, `dyd/secrets/`.
- Dependencies are declared under `dyd/requirements/` (files contain `root:<relative-path>` and are managed by `dryad root requirement add/remove`).

## Dryad concepts - Scopes
- A scope is a named execution context for scripts/aliases, used to partition commands in a garden.
- Scopes live under `dyd/shed/scopes/<scope-name>/`.
- Default scope is a symlink: `dyd/shed/scopes/default -> <scope-name>`.
- `dryad run` / `dryad script run` use `--scope=<name>` if provided; otherwise default scope.
- If no scope is set, these commands fail with “no scope set”.
- Scripts resolve to `dyd/shed/scopes/<scope>/script-run-<command>` and are executed directly.

## Dryad concepts - Stems
- A stem is a built artifact stored in the heap, marked by `dyd/type` with content `stem`.
- Entry point: `dyd/commands/dyd-stem-run` (or `--override` for run commands).
- Assets and metadata live under `dyd/assets/` and `dyd/traits/`.
- Dependencies are linked under `dyd/dependencies/`.
- Fingerprints are stored in `dyd/fingerprint`.

## Dryad concepts - Sprouts
- Sprouts are workspace-visible links to stems under `dyd/sprouts/...`.
- `dryad sprout run` executes the stem in its run environment.

## Dryad concepts - Heap
- The heap is a content-addressed store for stems and contexts: `dyd/heap/...`.
- It can be deleted and rebuilt; it is not source of truth.

## Dryad concepts - Contexts
- Contexts provide isolated HOME directories under `dyd/heap/contexts/<name>`.
- `dryad stem run` / `dryad sprout run` set `HOME` and `DYD_CONTEXT` to the context path.

## Dryad concepts - Shed
- `dyd/shed/` stores workspace configuration, including scopes and settings.
