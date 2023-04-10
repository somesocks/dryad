---
title: dryad roots
layout: default
parent: CLI
grand_parent: Usage
---

```
$ dryad root path --help
dryad root path [path]

Description:
    return the base path of the current root

Arguments:
    path   the path to start searching for a root at. defaults to current directory, optional
```dryad roots

Description:
    commands to work with dryad roots

Sub-commands:
    dryad roots list   list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)
    dryad roots path   return the path of the roots dir
