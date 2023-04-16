---
title: dryad run
layout: default
parent: CLI
grand_parent: Usage
---

# dryad run

```
$ dryad run --help
dryad run [--scope=string] [--inherit (default true)] <command> [-- args]

Description:
    alias for `dryad script run`

Arguments:
    command                        alias command
    -- args                        args to pass to the command, optional

Options:
        --scope                    set the scope for the command
        --inherit (default true)   pass all environment variables from the parent environment to the alias to exec
```