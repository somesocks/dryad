---
title: dryad script get
layout: default
parent: CLI
grand_parent: Usage
---

# dryad script get

```
$ dryad script get --help
dryad script get [--scope=string] [--log-level=string] [--log-format=string] [--help] <command>

Description:
    print the contents of a script

Arguments:
    command            the script name

Options:
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```