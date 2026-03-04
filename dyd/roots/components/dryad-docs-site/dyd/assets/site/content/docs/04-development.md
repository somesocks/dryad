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

## Diagnostics Engine

Dryad supports runtime diagnostics rules through the `DYD_DG` environment variable.

`DYD_DG` supports:

- `file:/absolute/path/to/diagnostics.yaml`
- `json:{...}`

### Fault Injection Example

This example injects an `EMLINK` pre-error on every `os.link` call:

```sh
DYD_DG='json:{"version":1,"seed":1,"rules":[{"id":"inject-emlink","op":"os.link","key":"*","when":{"mode":"every_n","count":1},"action":{"type":"error","error":"EMLINK"}}]}' dryad root build dyd/roots/root-01
```

For post-error behavior, set `action.phase` to `"post"`:

```json
{
  "type": "error",
  "phase": "post",
  "error": "EMLINK"
}
```

### Metrics Example

Diagnostics metrics can be configured per operation:

```yaml
version: 1
seed: 1
metrics:
  - id: os-link-metrics
    op: os.link
    output: stderr
    capture:
      calls: true
      errors: true
      timing: true
      sample_percent: 50
```

`sample_percent` is compiled to an internal power-of-two `1-in-N` sampler.
The emitted metrics include `sample_every` so the effective capture rate is explicit.

Example emitted line:

```json
{"point":"os.link","calls":21,"errors":0,"total_nanos":307521,"min_nanos":3814,"max_nanos":176874,"avg_nanos":14643,"sample_every":2}
```

Quick local check:

```sh
DYD_DG='json:{"version":1,"seed":1,"metrics":[{"id":"m-os-link","op":"os.link","output":"stderr","capture":{"calls":true,"errors":true,"timing":true,"sample_percent":100}}]}' \
dryad root build dyd/roots/root-01 --log-level=warn > /tmp/dyd-build.out 2> /tmp/dyd-build.err

grep -F '"point":"os.link"' /tmp/dyd-build.err
```

Diagnostics-focused E2E roots for dryad CLI live under:

- `dyd/roots/components/dryad-cli/tests/diagnostics-01--pre-error-fails-build`
- `dyd/roots/components/dryad-cli/tests/diagnostics-02--metrics-sampling`
- `dyd/roots/components/dryad-cli/tests/diagnostics-03--post-error-fails-build`
- `dyd/roots/components/dryad-cli/tests/diagnostics-04--metrics-output-streams`
- `dyd/roots/components/dryad-cli/tests/diagnostics-06--post-error-preserves-side-effects`

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
