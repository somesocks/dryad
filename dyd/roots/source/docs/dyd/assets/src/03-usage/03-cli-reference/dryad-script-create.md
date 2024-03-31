# dryad script create

```
$ dryad script create --help
dryad script create [--editor=string] [--scope=string] [--log-level=string] [--log-format=string] [--help] <command>

Description:
    create and edit a script

Arguments:
    command            the script name

Options:
        --editor       set the editor to use
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```