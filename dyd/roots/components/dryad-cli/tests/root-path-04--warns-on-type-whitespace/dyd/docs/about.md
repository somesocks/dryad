# root-path-04--warns-on-type-whitespace

This test case verifies that `dryad root path` tolerates whitespace in
`dyd/type` files and still returns the correct root path for each root.
It also verifies root search does not mutate source sentinels and logs
warnings with relative sentinel paths.

Whitespace cases covered:
- leading space
- trailing space
- leading newline
- trailing newline
- leading/trailing tab
