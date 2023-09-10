# dryad root replace

```
$ dryad root replace --help
dryad root replace [--log-level=string] [--log-format=string] [--help] <source> <replacement>

Description:
    replace all references to one root with references to another

Arguments:
    source             path to the source root
    replacement        path to the replacement root

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```