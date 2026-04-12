# sprouts-run-03--qualified-stdin-refs

Ensures `dryad sprouts run --from-stdin` accepts filesystem-qualified and
URL-qualified sprout refs, runs exactly the matching sprout variants, and
fails if a selector is specified in both a stdin sprout ref and `--variant`.
