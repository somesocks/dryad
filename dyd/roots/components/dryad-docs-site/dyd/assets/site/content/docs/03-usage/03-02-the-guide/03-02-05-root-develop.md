---
title: "03.02.05 - Root development"
description: "Working inside a root development environment."
type: docs
layout: single
---

# Root development

When you want to work on a root without mutating the source tree directly, you can use a root development environment.  This creates a temporary workspace with a snapshot of the root, then runs your editor inside that workspace.

To start a development session for the server root:

```
$ dryad root develop start dyd/roots/server --editor=sh
```

If you do not pass `--editor`, dryad will use `dyd/commands/dyd-root-develop` from the root if it exists, otherwise it falls back to `sh`.

The editor runs inside the dev workspace, so any changes you make there do not touch the source root until you save them.

The dev environment also includes the built dependencies of the root, so you can develop and test against the same toolchain and libraries used by the build.

## Status and save

From inside the dev environment, you can inspect and save changes:

```
$ dryad root develop status
$ dryad root develop save
```

Status output follows git porcelain v1 style codes, for example:

```
 M dyd/assets/main.go
 D dyd/assets/old.txt
?? dyd/assets/new.txt
```

`save` syncs the workspace changes back to the source root.  It respects `.dyd-ignore` and will report conflicts if the source root changed since the snapshot.

## Snapshot and reset

You can update the snapshot used by the dev environment:

```
$ dryad root develop snapshot
```

This packs the current workspace into the heap and updates the snapshot fingerprint.  A confirmation message is printed when the snapshot is saved.

To reset the workspace back to the snapshot state:

```
$ dryad root develop reset
```

`reset` overlays the workspace with the snapshot state.  It restores tracked files, and it leaves extra files in place.

## Stopping the editor

To stop the development session:

```
$ dryad root develop stop
```

If the root provides `dyd/commands/dyd-root-develop-start` or `dyd/commands/dyd-root-develop-stop`, those hooks run before and after the editor session.

## Notes

`dryad develop` is an alias for `dryad root develop`, so `dryad develop status` works as well.

Inside the dev environment, dryad sets `DYD_DEV_SOCKET` (path to the host IPC socket) and `DYD_CLI_BIN` (path to the host dryad binary).
