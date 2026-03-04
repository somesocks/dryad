# root-build-68--diagnostics-metrics-sampling

Ensures diagnostics metrics sampling applies deterministically to root-build `os.link` invocations, validating that `sample_percent=50` records exactly half (floor) of the `sample_percent=100` call count on identical fresh gardens.
