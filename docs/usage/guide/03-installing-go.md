---
title: 03 - Installing go
layout: default
nav_order: 2
parent: The Guide
grand_parent: Usage
---

# "Installing" go

We need the go toolchain and compiler in order to be able to compile our server into a binary.  However, we don't want to install go globally, that adds noise to our system, lowers reproducibility, and makes things like freezing the toolchain version more difficult.

Instead, we can create a root to "install" go as a package in our workspace, so that we can use our specific go version for the project.

