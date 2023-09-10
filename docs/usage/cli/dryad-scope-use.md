---
title: dryad scope use
layout: default
parent: CLI
grand_parent: Usage
---

# dryad scope use

```
$ dryad scope use --help
dryad scope use [--log-level=string] [--log-format=string] [--help] <name>

Description:
    set a scope to be active. alias for `dryad scopes default set`

Arguments:
    name               the name of the scope to set as active. use 'none' to unset the active scope

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```