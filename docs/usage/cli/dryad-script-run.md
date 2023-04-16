---
title: dryad script run
layout: default
parent: CLI
grand_parent: Usage
---

# dryad script run

```
$ dryad script run --help
dryad script run [--scope=string] [--inherit (default true)] <command> [-- args]

Description:
    run a script in the current scope

Arguments:
    command                        the script name
    -- args                        args to pass to the script, optional

Options:
        --scope                    set the scope for the command
        --inherit (default true)   pass all environment variables from the parent environment to the alias to exec
```