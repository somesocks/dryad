---
title: dryad sprouts exec
layout: default
parent: CLI
grand_parent: Usage
---

# dryad sprouts exec

```
$ dryad sprouts exec --help
dryad sprouts exec [--include] [--exclude] [--context=string] [--inherit] [--confirm] [--ignore-errors] [--scope=string] [-- args]

Description:
    execute each sprout in the current garden

Arguments:
    -- args               args to pass to each sprout on execution, optional

Options:
        --include         choose which sprouts are included
        --exclude         choose which sprouts are excluded
        --context         name of the execution context. the HOME env var is set to the path for this context
        --inherit         pass all environment variables from the parent environment to the stem
        --confirm         display the list of sprouts to exec, and ask for confirmation
        --ignore-errors   continue running even if a sprout returns an error
        --scope           set the scope for the command
```