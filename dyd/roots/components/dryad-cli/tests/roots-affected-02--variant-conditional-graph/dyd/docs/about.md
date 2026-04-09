# roots-affected-02--variant-conditional-graph

Checks that `dryad roots affected` reports exact affected variant refs, propagates through the variant dependency graph, handles no-variant roots, covers the selectable root path families, honors plain path selectors and conditional requirement files, treats root-global files and `dyd/variants` changes as affecting all variants of the owning root, and deduplicates mixed stdin unions.
