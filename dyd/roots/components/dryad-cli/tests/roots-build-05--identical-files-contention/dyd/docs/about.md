
# roots-build-05--identical-files-contention

This test case tests that `dryad roots build` runs successfully against a garden with a root with a large number of identical files. The identical files are intended to cause high thread contention on a single file in the heap.