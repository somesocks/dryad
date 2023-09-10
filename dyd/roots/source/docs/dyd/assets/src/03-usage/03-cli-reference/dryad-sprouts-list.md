# dryad sprouts list

```
$ dryad sprouts list --help
dryad sprouts list [--include] [--exclude] [--scope=string] [--log-level=string] [--log-format=string] [--help]

Description:
    list all sprouts of the current garden

Options:
        --include      choose which sprouts are included in the list
        --exclude      choose which sprouts are excluded from the list
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```