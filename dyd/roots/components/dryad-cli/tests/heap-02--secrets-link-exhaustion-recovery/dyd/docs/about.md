# heap-02--secrets-link-exhaustion-recovery

Ensures diagnostics `every_x_per_key` on `os.link` can simulate repeated
secret heap-link exhaustion, and that root builds recover from injected
`EMLINK` while preserving canonical secret-file sharing.
