# dryad root unlink

```
$ dryad root unlink --help
dryad root unlink [--log-level=string] [--log-format=string] [--help] <path>

Description:
    remove a dependency linked to the current root

Arguments:
    path               path to the dependency to unlink

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```