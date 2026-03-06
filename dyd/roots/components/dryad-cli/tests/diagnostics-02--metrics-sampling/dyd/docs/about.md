# diagnostics-02--metrics-sampling

Ensures diagnostics metrics sampling applies deterministically to root-build `os.link` invocations, validating that `when.x=2` records exactly half (floor) of the `when.x=1` call count on identical fresh gardens.
