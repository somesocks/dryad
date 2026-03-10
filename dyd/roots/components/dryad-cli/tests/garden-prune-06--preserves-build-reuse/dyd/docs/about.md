# garden-prune-06--preserves-build-reuse

This test case verifies that `dryad garden prune` does not cause a later `dryad root build dyd/roots/root-top` to rebuild its already-built helper dependency.
