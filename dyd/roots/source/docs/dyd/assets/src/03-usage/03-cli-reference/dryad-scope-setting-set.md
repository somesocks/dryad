# dryad scope setting set

```
$ dryad scope setting set --help
dryad scope setting set [--log-level=string] [--log-format=string] [--help] <scope> <setting> <value>

Description:
    set the value of a setting in a scope

Arguments:
    scope              the name of the scope
    setting            the name of the setting
    value              the new value for the setting

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```