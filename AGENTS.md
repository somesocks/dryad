# AGENTS.md

## Overview

- This repo is a Dryad garden.

## Dryad

- Dryad is a package manager / monorepo build tool designed for complex, multi-language package trees.
- Working with a dryad garden is done primarily through the `dryad` CLI tool.
  - Run `dryad` to see the current dryad version.
  - Run `dryad system commands` to see the complete list of dryad commands.
  - Run `dryad <command> --help` to see the help text for a command.
  - Dryad commands typically follow the convention `dryad <resource> <action> <arguments>`. `<resource>` can be singular or plural depending on whether you want to act on a single resource or a collection of resources. For example `dryad roots` has different actions than `dryad root`.
  - Most commands have a `--parallel=<int>` flag specifying the number of concurrent operations. Use `--parallel=1` to force serial execution.
  - Most commands have a `--log-level=<level>` flag. Use this to specify the logging level.


### Dryad - getting started

- Run `dryad scopes list` to see scopes in a garden
- Run `dryad roots list --scope=none` to see all roots in a garden
- Run `dryad scripts list --scope=<name>` to see all scripts for a scope
- Run `dryad roots graph` to see the package graph
- Run `dryad root ancestors <root_path>` to see the dependencies of a root (directly and indirectly).
- Run `dryad root descendants <root_path>` to see what depends on a root (directly and indirectly).


### Dryad concepts - gardens

- A _garden_ is a dryad package workspace. It has a well-defined filesystem structure:
	- `dyd/type` is a sentinel file containing the text `garden` (no newline). This indicates that this is a dryad garden.
	- `dyd/roots` is the collection of all the source packages. A `root` is source for a package to be built.
	- `dyd/heap` is where all build artifacts live. This should not be version-controlled.
	- `dyd/sprouts` contains symlinks to stems (built packages), following the same directory structure as the roots. If `dyd/roots/foo` is a root, then the built version of that root can be found at `dyd/sprouts/foo`. This should not be version-controlled.
	- `dyd/shed` contains configurations for the garden.
- Common garden commands include:
	- `dryad garden create` - create a new blank garden.
	- `dryad garden path` - return the base path for a garden (the parent garden of the current working directory).
	- `dryad garden prune` - remove unused build artifacts from the heap.
	- `dryad garden wipe` - remove all build artifacts from the heap.


### Dryad concepts - the heap

- The _heap_ is where all build artifacts in the garden live. It has a well-defined filesystem structure:
	- `dyd/heap/files` - content-addressed file store, each file is named for its fingerprint.
	- `dyd/heap/stems` - content-addressed package store. A _stem_ is a built package. Each stem in the heap is a directory named for its fingerprint, with hardlinks to heap files.
	- `dyd/heap/secrets` - a secondary content-addressed file store for heap files marked as secrets.
	- `dyd/heap/derivations` - a cache layer linking source fingerprints to build fingerprints, used for fast rebuild checks.
	- `dyd/heap/contexts` - a collection of "execution context" directories, or disposable home directories used during package builds and package runs. dryad replaces the home directory during executions to prevent home dir pollution.


### Dryad concepts - the shed

- The _shed_ is a configuration directory for the garden. It has a well-defined filesystem structure:
	- `dyd/shed/scopes` - where scopes are stored.


### Dryad concepts - scopes

- _scopes_ are virtual workspaces within a garden. You can use scopes to select "slices" of the garden to work with, or to group related commands.
- Scopes can be listed with `dryad scopes list`.
- A scope can be selected to be active by running `dryad scope use <XXX>`. The active scope can be seen by running `dryad scope active`.
- Almost all commands have a `--scope` flag. For example `dryad roots build --scope=dev` runs `dryad roots build` under the `dev` scope.
- `none` is a special scope keyword to run a command without any scope, even if a scope is active. `dryad roots build --scope=none` is equivalent to running `dryad roots build` without any scope active.
- Each scope is a directory under `dyd/shed/scopes`. Each scope has a well-defined filesystem structure:
	- `dyd/shed/scopes/<XXX>/.oneline` - The one-line description for the scope.
	- `dyd/shed/scopes/<XXX>/<command>` - Command-line arguments attached to a command. If a scope is active, these arguments are added to that command on its execution. For example, if `dyd/shed/scopes/dev/roots-build` file contained `--log-level=debug`, then running `dryad roots build --scope=dev` would rewrite the CLI arguments to `dryad roots build --scope=none --log-level=debug` before execution.
	- `dyd/shed/scopes/default` is a symlink to the active scope directory. It should not be version-controlled.
- _scripts_ can be added to a scope to extend dryad with custom commands. Scripts are custom shell scripts stored under `dyd/shed/scopes/<scope_name>/script-run-<script_name>`.
	- Scripts are run with `dryad run <script_name>`. Arguments can be passed to a script by using `--` escaping: `dryad run foo -- 1 2 3`.
	- The script has access to a DYD_SCOPE env var, specifying the current scope. This is so that the script can correctly pass the scope to subsequent `dryad` calls in the script.
	- The script has access to a DYD_LOG_LEVEL env var, specifying the current log level. This is so that the script can set the log level for subsequent `dryad` calls.


### Dryad concepts - roots

- _roots_ are source packages within a garden. All roots are stored under `dyd/roots/`.
- Each root is a directory with a well-defined filesystem structure:
  - `<root>/dyd/type` - a sentinel file containing the text `root` (no newline). This indicates that a directory is a root.
  - `<root>/dyd/assets` - _assets_ - source files for the root.
  - `<root>/dyd/commands` - _commands_ - build scripts or other commands for the root.
    - `<root>/dyd/commands/dyd-root-build` - this is the build script for the package.
  - `<root>/dyd/traits` - _traits_ - metadata files specifying traits for the root (name, version, license, etc.).
  - `<root>/dyd/docs` - _docs_ - documentation files for the root.
  - `<root>/dyd/secrets` - _secrets_ - secret assets for the root (deployment secrets, signing keys, etc.).
  - `<root>/dyd/requirements` - _requirements_ - the specification for dependencies of the root.
    - each requirement is a single file containing the relative path to another root in URL format (no newline), like so: `root:../../../foo`
    - the name of the requirement can be used to alias the name of a dependency. If `<root>/dyd/requirements/bar` has the URL `root:../../../foo`, then the built `foo` root will be linked in under the name `bar`.
- The root build process is a two-stage process.
  - First, the source of the root is packed into a _stem_ (a build package), stored into the heap.
    - This part of the process also converts requirements into dependencies, and links them to the stem under `<stem>/dyd/dependencies/<name>`
  - Second, the build stem is executed in a disposable build environment to build the resulting package
    - This build environment is given a DYD_BUILD env var to specify the destination location for the package.
    - The build script (`dyd/commands/dyd-root-build`) should create a new stem at the path in DYD_BUILD.
  - Afterwards the built stem is packed into the heap, and then linked as a sprout under `<garden>/dyd/sprouts/<root_name>`
  - This two-stage build process enforces correctness and improves caching behavior.
- Common root commands include:
  - `dryad roots build` - build all roots
    - `--parallel=1` - run a serial build
    - `--join-stdout --join-stderr` - log the stdout and stderr of child processes to the shell
    - `--log-level=debug` - for increased logs
  - `dryad roots list` - list all roots
  - `dryad roots graph` - see the complete package graph
  - `dryad root create <path>` - create an empty root template at <path>
  - `dryad root build <path?>` - build the root specifically at path. If <path> is not provided, try the current working directory
    - `--join-stdout --join-stderr` - log the stdout and stderr of child processes to the shell
    - `--log-level=debug` - for increased logs
  - `dryad root copy <source> <destination>` - copy a dryad root to a new location, while keeping requirements pointing at the original dependencies.
    -  `--unpin` - treat requirements as "floating" dependencies. If a requirement exists relative to the new location use it, otherwise use the original.
    - Warning: you cannot copy a root inside another root.
  - `dryad root move <source> <destination>` - move a root to a new location, while keeping requirements pointing to the original dependencies AND updating all other roots to point to the new location.
    -  `--unpin` - treat requirements as "floating" dependencies. If a requirement exists relative to the new location use it, otherwise use the original.
    - Warning: you cannot move a root inside another root.
  - `dryad root develop start <source>` - this creates an interactive development environment for a root by performing the first stage of the build, and then dropping you into a shell in the disposable build environment used to build a stem from the source. Useful for debugging or interactive development.
    - Interactive developments are complex, you should read all `dryad root develop` sub-command help text before you use it.


### Dryad concepts - stems

- _stems_ are content-addressed built packages stored in `dyd/heap/stems`.
- A stem is an immutable, fingerprinted artifact produced from a root source state.
- During packing/build prep, root requirements are linked into the stem under `<stem>/dyd/dependencies/<name>`.


### Dryad concepts - sprouts

- _sprouts_ are generated links to stems.
- `dyd/sprouts/<path>` mirrors `dyd/roots/<path>` for built outputs.
- Sprouts are build artifacts, not source code: do not edit them directly.
- `dyd/sprouts` should not be version-controlled.
- If sprouts are stale or missing, regenerate them by rebuilding roots (for example `dryad roots build` or `dryad root build <path>`).
- Common sprout commands include:
  - `dryad sprouts list` - list all sprouts in the current garden.
  - `dryad sprouts path` - print the sprouts directory path.
  - `dryad sprouts prune` - synchronize sprouts directory structure with roots.
  - `dryad sprouts run -- -- <args>` - run each sprout in the garden.
  - `dryad sprout run <path> -- -- <args>` - run a single sprout.
