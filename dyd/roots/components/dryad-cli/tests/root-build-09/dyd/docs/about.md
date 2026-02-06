# root-build-09

This test verifies that built stems keep `dyd/requirements` as manifest files and
set up `dyd/dependencies` as stem symlinks.

Coverage:
- `dyd/requirements/root-02` stays as plain-text `root:../../../root-02`
- `dyd/dependencies/root-02` is a symlink to a built dependency stem
- dependency stem path basename matches its `dyd/fingerprint`
