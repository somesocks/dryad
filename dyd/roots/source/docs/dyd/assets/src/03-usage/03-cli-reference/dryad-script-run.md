# dryad script run

```
$ dryad script run --help
dryad script run [--inherit] [--scope=string] [--log-level=string] [--log-format=string] [--help] <command> [-- args]

Description:
    run a script in the current scope

Arguments:
    command            the script name
    -- args            args to pass to the script, optional

Options:
        --inherit      pass all environment variables from the parent environment to the alias to exec
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```