
# roots-build-10--move-root-into-old-path

This test case verifies that `dryad roots build` still succeeds when a root that was previously built at `dyd/roots/root-01` is moved to `dyd/roots/root-01/1.0` and rebuilt. It specifically checks that `dyd/sprouts/root-01` is converted from a symlink into a directory so `dyd/sprouts/root-01/1.0` can be linked.
