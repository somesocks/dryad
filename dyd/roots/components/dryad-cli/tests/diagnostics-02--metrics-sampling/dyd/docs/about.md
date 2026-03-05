# diagnostics-02--metrics-sampling

Ensures diagnostics metrics sampling applies deterministically to root-build `os.link` invocations, validating that `when.every_n=2` records exactly half (floor) of the `when.every_n=1` call count on identical fresh gardens.
