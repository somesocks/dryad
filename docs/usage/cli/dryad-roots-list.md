---
title: dryad roots list
layout: default
parent: CLI
grand_parent: Usage
---

# dryad roots list

```
$ dryad roots list --help
dryad roots list [--include] [--exclude] [--scope=string] [path]

Description:
    list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)

Arguments:
    path            path to the base root (or garden) to list roots in, optional

Options:
        --include   choose which roots are included in the list
        --exclude   choose which roots are excluded from the list
        --scope     set the scope for the command
```