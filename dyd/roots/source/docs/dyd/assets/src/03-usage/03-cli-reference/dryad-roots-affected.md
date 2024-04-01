# dryad roots affected

```
$ dryad roots affected --help
dryad roots affected [--relative=boolean] [--scope=string] [--log-level=string] [--log-format=string] [--help]

Description:
    take a list of files from stdin, and print a list of roots that may depend on those files

Options:
        --relative     print roots relative to the base garden path. default true
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```