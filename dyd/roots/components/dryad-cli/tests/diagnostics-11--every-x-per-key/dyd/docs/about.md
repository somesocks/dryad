# diagnostics-11--every-x-per-key

Ensures diagnostics `every_x_per_key` on `os.link` can simulate repeated heap-link exhaustion, and that root builds recover from injected `EMLINK` by materializing repeated content without requiring every destination to remain hard-linked.
