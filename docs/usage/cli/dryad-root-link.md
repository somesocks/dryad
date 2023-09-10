---
title: dryad root link
layout: default
parent: CLI
grand_parent: Usage
---

# dryad root link

```
$ dryad root link --help
dryad root link [--log-level=string] [--log-format=string] [--help] <path> [alias]

Description:
    link a root as a dependency of the current root

Arguments:
    path               path to the root you want to link as a dependency
    alias              the alias to link the root under. if not specified, this defaults to the basename of the linked root, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```