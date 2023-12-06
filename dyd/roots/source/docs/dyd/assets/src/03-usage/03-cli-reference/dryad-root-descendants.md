# dryad root descendants

```
$ dryad root descendants --help
dryad root descendants [--log-level=string] [--log-format=string] [--help] [root_path]

Description:
    list all roots that depend on the selected root (directly and indirectly)

Arguments:
    root_path          path to the root, optional

Options:
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```