---
title: "03.02.01 - Creating a garden"
description: "Initializing a dryad workspace."
type: docs
layout: single
---

# Creating a garden

Once dryad is installed, a garden can be created in the current directory by running `dryad garden init`.

This should create a file structure that looks something like:

```
./
  dyd/
    heap/
    roots/
    shed/
    sprouts/
    garden
```

In order to take advantage of the filesystem, dryad enforces a strict structure for workspaces and packages:
- gardens are contained in a `/dyd` directory
- gardens have a `heap` directory for caching build artifacts
- gardens have a `roots` directory for the source packages to be built
- gardens have a `shed` for storing garden configurations
- gardens have a `sprouts` directory, where built packages are linked
- gardens have an empty `/dyd/type` file at the base with the content `garden` (to flag the directory as a garden)

You can read more about the heap, roots, the shed, and sprouts in the [Concepts]({{ site.baseurl }}{% link concepts/index.md %}).  But, the heap and the sprouts are build artifacts (or links to build artifacts).  So, if this project is going to be stored in a version control system, you should likely ignore them.

Here is an example gitignore for ignoring the heap and sprouts in a git project:

```
# gitignore paths for a dryad garden
/**/dyd/heap/
/**/dyd/sprouts/
```

