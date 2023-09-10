# dryad scope setting unset

```
$ dryad scope setting unset --help
dryad scope setting unset [--log-level=string] [--log-format=string] [--help] <scope> <setting>

Description:
    remove a setting from a scope

Arguments:
    scope              the name of the scope
    setting            the name of the setting

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```