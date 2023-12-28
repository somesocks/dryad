# dryad roots owning

```
$ dryad roots owning --help
dryad roots owning [--scope=string] [--log-level=string] [--log-format=string] [--help]

Description:
    list all roots that are owners of the provided files. The files to check should be provided as relative or absolute paths through stdin.

Options:
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```