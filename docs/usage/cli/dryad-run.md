---
title: dryad run
layout: default
parent: CLI
grand_parent: Usage
---

# dryad run

```
$ dryad run --help
dryad run [--scope=string] [--inherit] [--log-level=string] [--log-format=string] [--help] <command> [-- args]

Description:
    alias for `dryad script run`

Arguments:
    command            alias command
    -- args            args to pass to the command, optional

Options:
        --scope        set the scope for the command
        --inherit      pass all environment variables from the parent environment to the alias to exec
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```