# dryad secrets path

```
$ dryad secrets path --help
dryad secrets path [--log-level=string] [--log-format=string] [--help] <path>

Description:
    print the path to the secrets for the current package, if it exists

Arguments:
    path               path to the stem base dir

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```