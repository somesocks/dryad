# dryad stem run

```
$ dryad stem run --help
dryad stem run [--context=string] [--inherit] [--override=string] [--confirm=string] [--log-level=string] [--log-format=string] [--help] <path> [-- args]

Description:
    execute the main for a stem

Arguments:
    path               path to the stem base dir
    -- args            args to pass to the stem, optional

Options:
        --context      name of the execution context. the HOME env var is set to the path for this context
        --inherit      pass all environment variables from the parent environment to the stem
        --override     run this executable in the stem run envinronment instead of the main
        --confirm      ask for a confirmation string to be entered to execute this command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```