---
title: dryad scope setting get
layout: default
parent: CLI
grand_parent: Usage
---

# dryad scope setting get

```
$ dryad scope setting get --help
dryad scope setting get [--log-level=string] [--log-format=string] [--help] <scope> <setting>

Description:
    print the value of a setting in a scope, if it exists

Arguments:
    scope              the name of the scope
    setting            the name of the setting

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```