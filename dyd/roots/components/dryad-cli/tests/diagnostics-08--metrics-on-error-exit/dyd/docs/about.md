# diagnostics-08--metrics-on-error-exit

Ensures diagnostics metrics are emitted on process exit when a command returns a non-zero code (running outside a garden), by collecting `os.read_file` metrics to stdout.
