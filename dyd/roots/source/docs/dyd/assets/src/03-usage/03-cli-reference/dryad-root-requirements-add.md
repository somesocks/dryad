# dryad root requirements add

```
$ dryad root requirements add --help
dryad root requirements add [--log-level=string] [--log-format=string] [--help] <path> [alias]

Description:
    add a root as a dependency of the current root

Arguments:
    path               path to the root you want to add as a dependency
    alias              the alias to add the root under. if not specified, this defaults to the basename of the linked root, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```