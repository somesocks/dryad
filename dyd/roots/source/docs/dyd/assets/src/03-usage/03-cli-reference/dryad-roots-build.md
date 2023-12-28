# dryad roots build

```
$ dryad roots build --help
dryad roots build [--path=string] [--include=multi-string] [--exclude=multi-string] [--scope=string] [--log-level=string] [--log-format=string] [--help]

Description:
    build selected roots in a garden

Options:
        --path         the target path for the garden to build
        --include      choose which roots are included in the build
        --exclude      choose which roots are excluded from the build
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```