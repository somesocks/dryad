# heap-03--files-link-exhaustion-parallel-contention

Validates that concurrent builds refresh the canonical heap file in place when
the preseeded canonical inode returns `EMLINK` during materialization.
