# heap-04--canonical-refresh-prune-lifecycle

Ensures a canonical heap file refreshed after injected `EMLINK` stays live through later builds, while `dryad garden prune` removes the unreachable old stem that still referenced the pre-refresh inode.
