# dryad scripts list

```
$ dryad scripts list --help
dryad scripts list [--path] [--oneline] [--scope=string] [--log-level=string] [--log-format=string] [--help]

Description:
    list all available scripts in the current scope

Options:
        --path         print the path to the scripts instead of the script run command
        --oneline      print the oneline decription of each command
        --scope        set the scope for the command
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```