---
title: "04 - Development"
description: "How to build/develop/modify dryad itself."
type: docs
layout: single
---

# Development

## Tools

The dryad repository is a dryad garden itself, so to develop dryad you need a working copy of dryad  installed on your machine.

You can either install a previous dryad release, or you can use a "bootstrap" build - building a native copy of dryad.

### Previous releases

Install instructions for previous releases can be found on [03.01 - Installing](../03-usage/03-01-installing.html)

### Bootstrap build

To build dryad directly, you need to have a copy of the golang toolchain installed on your machine. Once go is installed, you can run the shell script `utils/bootstrap-build.sh` to directly build the dryad cli tool. The resulting binary should be placed `bootstrap/dryad`. You can either copy this binary somewhere into your system path, or you can add the bootstrap directory to your path temporarily.

A [direnv](https://direnv.net/) .envrc file is included in the repo, that adds the bootstrap directory to your path and uses [nix](https://nixos.org/) to install a copy of go when you enter the repo directory in a shell.

## Development Scopes

Once you have a working copy of dryad installed, you can use one of several scopes to work with development. Note that the dryad repo has roots to install any other tools needed for the build, so you should not need to install any other system tools.

```
$ dryad scopes list
dev - scope for developing dryad
docs - scope for developing the docs site
release - scope for building and publishing releases
```

### The dev scope

The dev scope is used for developing dryad itself. It has several commands:

```
$ dryad scope use dev
$ dryad scripts list
dryad run build - build dryad
dryad run test - run test cases
```

Running `dryad run build` will build a machine-native version of dryad, and all test case roots.

Running `dryad run test` will run all test cases against the native dryad root.

### The docs scope

The docs scope is used for developing and updating this docs site.

The docs site is a [Hugo](https://gohugo.io/) static site using the [Lotus Docs](https://lotusdocs.dev/) theme.

```
$ dryad scope use docs
$ dryad scripts list
dryad run build - build the docs site
dryad run open - open the compiled docs in the browser
```

Running `dryad run build` will build the docs site.

Running `dryad run open` will start a Caddy web server to host the docs site locally, and open the site in a browser window.

### The release scope

The release scope is used for building release versions of dryad for all supported platforms, and publishing them.

```
$ dryad scope use release
$ dryad scripts list
dryad run build - runs a build of the garden that includes all release builds
dryad run publish-release-tag - creates a git tag for the current release version, and pushes it to remote
```