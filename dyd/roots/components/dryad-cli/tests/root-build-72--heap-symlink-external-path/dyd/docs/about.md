# root-build-72--heap-symlink-external-path

Validates the detached-heap lifecycle when `dyd/heap` is a symlink to an external heap directory: `root build`, `sprout run`, `garden prune`, and `garden wipe` all succeed, and rebuilds still repopulate the external heap afterward.
