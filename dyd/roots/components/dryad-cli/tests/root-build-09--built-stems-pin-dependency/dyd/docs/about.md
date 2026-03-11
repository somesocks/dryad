# root-build-09--built-stems-pin-dependency

This test verifies that built stems pin dependency fingerprints into
`dyd/requirements` and set up `dyd/dependencies` as stem symlinks.

Coverage:
- `dyd/requirements/root-02` is a pinned fingerprint file
- `dyd/dependencies/root-02` is a symlink to a built dependency stem
- dependency stem path basename matches its `dyd/fingerprint`
- source requirement files tolerate trailing whitespace and are not rewritten
- malformed requirement whitespace logs a warning with a relative requirement path
