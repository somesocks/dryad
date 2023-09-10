# dryad root move

```
$ dryad root move --help
dryad root move [--log-level=string] [--log-format=string] [--help] <source> <destination>

Description:
    move a root to a new location and correct all references

Arguments:
    source             path to the source root
    destination        destination path for the root

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```