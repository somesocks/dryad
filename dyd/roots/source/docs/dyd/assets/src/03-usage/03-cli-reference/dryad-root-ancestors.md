# dryad root ancestors

```
$ dryad root ancestors --help
dryad root ancestors [--relative=boolean] [--log-level=string] [--log-format=string] [--help] [root_path]

Description:
    list all roots the selected root depends on (directly and indirectly)

Arguments:
    root_path          path to the root, optional

Options:
        --relative     print roots relative to the base garden path. default true
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```