# Collection variant concurrency

Verifies overlap and a shared concurrency limit across sprouts and variants,
including qualified stdin targets and serial execution. Tests parallel error
continuation and serial fail-fast/ignore-errors behavior at sprout granularity.
