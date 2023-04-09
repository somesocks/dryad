---
title: Usage
layout: default
nav_order: 4
---

# Usage

All commands in the `dryad` cli tool follow the same convention:
`dryad <RESOURCE> <ACTION> <ARGUMENTS>`.  Running `dryad` without any arguments prints help text and a list of commands.

Here is a list of dryad commands:

NOTE: Commands may be added or changed frequently during development, and this list may not be up to date.

- `dryad garden build` - build all roots in the garden
- `dryad garden init` - initialize a garden
- `dryad garden pack` - pack the current garden into an archive 
- `dryad garden path` - return the base path for a garden
- `dryad garden prune` - clear all build artifacts out of the garden not actively linked to a sprout or a root
- `dryad garden wipe` - clear all build artifacts out of the garden

- `dryad root add` - add a root as a dependency of the current root
- `dryad root build` - build a specified root
- `dryad root init` - create a new root directory structure in the current dir
- `dryad root path` - return the base path of the current root

- `dryad roots list` - list all roots that are dependencies for the current root
- `dryad roots path` - return the path of the roots dir

- `dryad secrets fingerprint` - calculate the fingerprint for the secrets in a stem/root
- `dryad secrets list` - list the secret files in a stem/root
- `dryad secrets path` - print the path to the secrets for the current package, if it exists

- `dryad stem exec` - execute the main for a stem
- `dryad stem fingerprint` - calculate the fingerprint for a stem dir
- `dryad stem files` - list the files in a stem
- `dryad stem pack` - pack the stem at the target path into a tar archive
- `dryad stem path` - return the base path of the current root
- `dryad stem unpack` - unpack a stem archive at the target path - `and import it into the current garden

- `dryad stems list` - list all stems that are dependencies for the current root
- `dryad stems path` - return the path of the stems dir

