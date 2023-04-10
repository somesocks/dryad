---
title: dryad stem exec
layout: default
parent: CLI
grand_parent: Usage
---

```
dryad stem exec [--execPath=string] [--context=string] [--inherit] <path> [-- args]

Description:
    execute the main for a stem

Arguments:
    path             path to the stem base dir
    -- args          args to pass to the stem, optional

Options:
        --execPath   path to the executable running `dryad stem exec`. used for path setting
        --context    name of the execution context. the HOME env var is set to the path for this context
        --inherit    pass all environment variables from the parent environment to the stem
```