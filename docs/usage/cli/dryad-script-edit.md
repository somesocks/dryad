---
title: dryad script edit
layout: default
parent: CLI
grand_parent: Usage
---

# dryad script edit

```
$ dryad script edit --help
dryad script edit [--editor=string] [--scope=string] [--log-level=string] [--log-format=string] [--help] <command>

Description:
    edit a script

Arguments:
    command            the script name

Options:
        --editor       set the editor to use
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```