# dryad root develop

```
$ dryad root develop --help
dryad root develop [--editor=string] [--arg=multi-string] [--inherit=boolean] [--scope=string] [--log-level=string] [--log-format=string] [--help] [path]

Description:
    create a temporary development environment for a root

Arguments:
    path               path to the root to develop, optional

Options:
        --editor       choose the editor to run in the root development environment
        --arg          argument to pass to the editor
        --inherit      inherit env variables from the host environment
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```