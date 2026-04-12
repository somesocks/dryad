# sprout-run-09--qualified-variant-ref

Ensures `dryad sprout run` accepts filesystem-qualified and URL-qualified
sprout refs, runs exactly the matching sprout variants, and fails if a
selector is specified in both the sprout ref and `--variant`.
