# heap-02--secrets-link-exhaustion-recovery

Ensures repeated canonical refreshes in `dyd/heap/secrets` recover from
injected `EMLINK`, rotate through multiple secret inodes over time, and still
prune away older unreachable stem generations cleanly.
