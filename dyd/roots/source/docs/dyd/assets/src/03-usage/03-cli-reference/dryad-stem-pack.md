# dryad stem pack

```
$ dryad stem pack --help
dryad stem pack [--log-level=string] [--log-format=string] [--help] <stemPath> [targetPath]

Description:
    pack the stem at the target path into a tar archive

Arguments:
    stemPath           the path to the stem to pack
    targetPath         the path (including name) to output the archive to, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```