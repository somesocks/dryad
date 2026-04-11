# root-build-23--variant-flag-selects-subset

Ensures `dryad root build` accepts `--variant`, filesystem-qualified root refs,
and URL-qualified root refs to select a subset of concrete variants, and rejects
selectors specified in both `root_ref` and `--variant`.
