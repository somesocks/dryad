---
title: "05 - Sprouts"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: true
# bookComments: false
# bookSearchExclude: false
---

# Sprouts

To track the built versions of each root, A garden will also have a **sprouts** directory at `dyd/sprouts/`.

The sprouts directory will automatically be created with the same filesystem structure as the roots directory during the build process.  So, a garden with two roots at `dyd/roots/tools/foo`, `dyd/roots/tests/foo-tests`, after a build will have two sprouts `dyd/sprouts/tools/foo` and `dyd/sprouts/tests/foo-tests` that link to the stems that resulted from the build.