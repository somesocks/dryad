---
title: dryad stem unpack
layout: default
parent: CLI
grand_parent: Usage
---

# dryad stem unpack

```
$ dryad stem unpack --help
dryad stem unpack [--log-level=string] [--log-format=string] [--help] <archive>

Description:
    unpack a stem archive at the target path and import it into the current garden

Arguments:
    archive            the path to the archive to unpack

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```