# dryad root copy

```
$ dryad root copy --help
dryad root copy [--log-level=string] [--log-format=string] [--help] <source> <destination>

Description:
    make a copy of the specified root at a new location

Arguments:
    source             path to the source root
    destination        destination path for the root copy

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```