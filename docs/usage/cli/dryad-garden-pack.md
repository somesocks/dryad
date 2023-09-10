---
title: dryad garden pack
layout: default
parent: CLI
grand_parent: Usage
---

# dryad garden pack

```
$ dryad garden pack --help
dryad garden pack [--log-level=string] [--log-format=string] [--help] [gardenPath] [targetPath]

Description:
    pack the current garden into an archive 

Arguments:
    gardenPath         the path to the garden to pack, optional
    targetPath         the path (including name) to output the archive to, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```