# root-path-04

This test case verifies that `dryad root path` tolerates whitespace in
`dyd/type` files and still returns the correct root path for each root.

Whitespace cases covered:
- leading space
- trailing space
- leading newline
- trailing newline
- leading/trailing tab
