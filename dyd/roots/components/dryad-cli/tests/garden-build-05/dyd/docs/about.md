
# garden-build-05

This test case tests that `dryad garden build` runs successfully against a garden with a root with a large number of identical files. The identical files are intended to cause high thread contention on a single file in the heap.