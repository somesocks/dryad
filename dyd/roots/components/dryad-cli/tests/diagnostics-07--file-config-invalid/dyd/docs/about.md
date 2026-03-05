# diagnostics-07--file-config-invalid

Ensures malformed diagnostics YAML provided via `DYD_DIAG=file:...` fails fast during diagnostics initialization with a parse error, and does not affect subsequent builds without diagnostics.
