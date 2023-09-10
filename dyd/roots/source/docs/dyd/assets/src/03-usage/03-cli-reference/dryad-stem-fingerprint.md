# dryad stem fingerprint

```
$ dryad stem fingerprint --help
dryad stem fingerprint [--exclude=string] [--log-level=string] [--log-format=string] [--help] [path]

Description:
    calculate the fingerprint for a stem dir

Arguments:
    path               path to the stem base dir, optional

Options:
        --exclude      a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory
        --log-level    set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'
        --log-format   set the logging format. can be one of 'console' or 'json'.  defaults to 'console'
        --help         display help text for this command
```