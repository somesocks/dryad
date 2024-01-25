# dryad root requirements remove

```
$ dryad root requirements remove --help
dryad root requirements remove [--log-level=string] [--log-format=string] [--help] <path>

Description:
    remove a requirement from the current root

Arguments:
    path               path to the dependency to remove

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```