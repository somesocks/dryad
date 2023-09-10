# dryad root path

```
$ dryad root path --help
dryad root path [--log-level=string] [--log-format=string] [--help] [path]

Description:
    return the base path of the current root

Arguments:
    path               the path to start searching for a root at. defaults to current directory, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```