# root-build-04

This test case verifies that `dryad root build` tolerates whitespace in
`dyd/type` files inside a fixture garden and sanitizes them during build.

Whitespace cases covered:
- leading space
- trailing space
- leading newline
- trailing newline
- leading/trailing tab
