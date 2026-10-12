# Sprout variant concurrency

Holds real built variant processes until release to prove default/configured
parallel overlap, serial execution, and the process limit. A failing variant must
not cause the CLI to return while an already-started sibling is blocked.
