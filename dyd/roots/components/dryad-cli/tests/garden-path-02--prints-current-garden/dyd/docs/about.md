
# garden-path-02--prints-current-garden

This test case tests that `dryad garden path` runs successfully against a garden,
and returns the correct path.
It also verifies that a whitespace-malformed garden sentinel (`dyd/type`)
is tolerated with a warning and is not rewritten.
