---
title: "02.09 - Contexts"
description: "Execution environments for running sprouts."
type: docs
layout: single
---

# Contexts

Many applications expect to be able to write state to a "global" store, usually in the home directory of the user.

In order to avoid global state pollution, dryad executes stems under an artifical home directory, a **context**.  Contexts are stored in the heap, at `dyd/heap/contexts`.

You can have multiple contexts, and select which context to use when executing a stem (the `--context` option).  The context is also a good place to pre-populate application state when initializing an application as well.