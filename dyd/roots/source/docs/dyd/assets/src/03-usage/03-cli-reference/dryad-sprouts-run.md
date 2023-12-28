# dryad sprouts run

```
$ dryad sprouts run --help
dryad sprouts run [--include] [--exclude] [--context=string] [--inherit] [--confirm=string] [--ignore-errors] [--scope=string] [--log-level=string] [--log-format=string] [--help] [-- args]

Description:
    run each sprout in the current garden

Arguments:
    -- args               args to pass to each sprout on execution, optional

Options:
        --include         choose which sprouts are included
        --exclude         choose which sprouts are excluded
        --context         name of the execution context. the HOME env var is set to the path for this context
        --inherit         pass all environment variables from the parent environment to the stem
        --confirm         ask for a confirmation string to be entered to execute this command
        --ignore-errors   continue running even if a sprout returns an error
        --scope           set the scope for the command
        --log-level       set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format      set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help            display help text for this command
```