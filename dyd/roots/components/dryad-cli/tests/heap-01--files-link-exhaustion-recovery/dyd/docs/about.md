# heap-01--files-link-exhaustion-recovery

Ensures diagnostics `every_x_per_key` on `os.link` can simulate repeated heap-link exhaustion, and that root builds recover from injected `EMLINK` while preserving canonical heap-file sharing for repeated content.
