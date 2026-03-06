# root-build-67--heap-files-rollover-parallel-contention

Validates that concurrent builds reuse a single heap file replica when the
canonical heap file returns `EMLINK` during materialization.
