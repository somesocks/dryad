# dryad
Source code for the dryad package manager.

## Design Goals

dryad is built to manage the development and publishing of complex package trees.  It was built with several goals in mind:

- **Hermetic builds** - Dryad packages should be stable across builds by default, and only depend on their source code and explicitly marked dependendencies.  Any unstable packages must be explicity marked as such.  All builds should happend in a (somewhat) isolated, read-only build environment.  All packages should be addressable by a content-hash-based fingerprint.  Hermetic builds enable accurate build caching, and easily reproducible installs.

- **Language agnostic** - Dryad packages have no special build language or configuation formats.  All package properties / assets should be organized in a standardized filesystem structure, to maximize compatibility with different tools and programming languages.

- **Registry agnostic** - The _only_ unique identifier for a dryad package is its fingerprint.  There should be no global package names, no authoritative package registry, no "namespaces", etc.  Package authenticity should be verified by cryptographic signatures independent of package distribution.  Package "upgrades" should be decided by heuristics on package traits.

## Concepts

### Gardens

A workspace for downloading/installing/developing packages is called a _garden_.  All packages belong to a single garden.

### The heap

Every garden has a _heap_ - where package assets are stored in a content-addressed filesystem structure.

The location of the heap within a garden should be `dyd/heap`.

### Stems

A dryad package is called a _stem_.  Every stem is stored in a read-only filesystem structure inside the heap.  Stems must always be treated as read-only, except during the build process.

Every stem has a fingerprint, a unique, content-hash-based id for that stem.  The fingerprint is formatted as `<algorithm>-<hash as hexcode>`, f.ex. `blake2b-b838157d9ff40dc52b4c832c1a3e50b8`.  The fingerprint for a stem is stored in a plain-text file located at `dyd/fingerprint`.

Every stem may have any number of code/library/data files, called _assets_.  All assets should be stored in the directory `dyd/assets/`.

Every stem may have any number of extra properties to describe the package, called _traits_.  All traits should be stored in the `dyd/traits/` directory for that stem.  The `dyd/traits/` directory should be treated as a key-value text store where possible, so traits should be stored individually in plain-text files `dyd/traits/` where it makes sence.

Every stem may have a _single_ executable file call the _main entrypoint_ or _main_.  This file should be located at `dyd/main`.  If multiple executables are packaged in a stem, this should be the entrypoint for invoking/passing commands to them.

Every stem may be dependent on any number of other stems.  These stems should be symlinked into the `dyd/stems/` directory.  Each stem is linked as a package dependency using a package _alias_.  A dependency with the alias `foo` should be symlinked to `dyd/stems/foo`.  The alias  for a dependency has _no relation_ to the dependency itself.  The same stem can be linked as a dependency twice under two different alias names, for example.

When a stem fingerprint is calculated, the _fingerprints_ and _traits_ of all dependencies are included as part of the fingerprint calculation.  This "fingerprint of fingerprints" is what creates a fully-reproducible package tree where every dependency is packaged to a specific version.

Likewise, when a stem is packaged into an isolated archive, the _fingerprints_ and _traits_ of all dependencies are considered part of that stem, and must also be included.

Every stem also has a collection of auto-generated _path stubs_ for each dependency present in `dyd/path/`.  These stubs are included as part of the executable path when `dyd/main` is invoked, which provides a (semi) isolated execution environment for each stem.  They are auto-generated during the stem build process, and included in calculation of the stem fingerprint.

If a stem needs access to credentials, keys, or other temporary secrets, they will be mounted in the `dyd/secrets/` directory at install or run-time.  The secrets a stem _must be available_ during the stem build process, and will be included (indirectly) as part of the stem fingerprint.  If a stem should have access to secrets at run-time, a _secrets fingerprint_ will be generated and stored in the stem at `dyd/secrets-fingerprint`.

### Roots

A _root_ is the package environment for building a stem.  All roots in a garden are stored under `dyd/roots/`.  Roots have a similar filesystem structure to stems, with `dyd/assets/`, `dyd/traits`, `dyd/secrets`, and a `dyd/main`, but no path or fingerprint.

Roots may have a `dyd/stems/` directory to specify dependencies.  In addition, roots may have a `dyd/roots/` directory, where other roots in the workspace may be symlinked as dependencies.

The root build process is roughly:

1. For a root, build all root dependencies, and get the fingerprints of the resulting stems to add as dependencies.
2. Naively package a root as a stem by bundling it, adding the root dependencies, fingerprinting it, and adding it into the heap.
3. Execute the main of that root, providing it with a temporary directory in which to build a stem.
4. Fingerprint the resulting stem, and add it into the heap.

The build process may be skipped if the root has already been built and the resulting stem already added to the heap.

### Sprouts

To track the built versions of each root, A garden will also have a _sprouts_ directory at `dyd/sprouts/`.  The sprouts directory will automatically be created with the same filesystem structure as the roots directory during the build process.  So, a garden with two roots at `dyd/roots/tools/foo`, `dyd/roots/tests/foo-tests`, after a build will have two sprouts `dyd/sprouts/tools/foo` and `dyd/sprouts/tests/foo-tests` that link to the stems that resulted from the build.

## Usage

All commands in the `dryad` cli tool follow the same convention:
`dryad <RESOURCE> <ACTION> <ARGUMENTS>`.  Running `dryad` without any arguments prints help text and a list of commands.

Here is a list of dryad commands:

- `dryad garden build` - build all roots in the garden
- `dryad garden init` - initialize a garden
- `dryad garden pack` - pack the current garden into an archive 
- `dryad garden path` - return the base path for a garden
- `dryad garden prune` - clear all build artifacts out of the garden not actively linked to a sprout or a root
- `dryad garden wipe` - clear all build artifacts out of the garden

- `dryad root add` - add a root as a dependency of the current root
- `dryad root build` - build a specified root
- `dryad root init` - create a new root directory structure in the current dir
- `dryad root path` - return the base path of the current root

- `dryad roots list` - list all roots that are dependencies for the current root
- `dryad roots path` - return the path of the roots dir

- `dryad secrets fingerprint` - calculate the fingerprint for the secrets in a stem/root
- `dryad secrets list` - list the secret files in a stem/root
- `dryad secrets path` - print the path to the secrets for the current package, if it exists

- `dryad stem exec` - execute the main for a stem
- `dryad stem fingerprint` - calculate the fingerprint for a stem dir
- `dryad stem files` - list the files in a stem
- `dryad stem pack` - pack the stem at the target path into a tar archive
- `dryad stem path` - return the base path of the current root
- `dryad stem unpack` - unpack a stem archive at the target path - `and import it into the current garden

- `dryad stems list` - list all stems that are dependencies for the current root
- `dryad stems path` - return the path of the stems dir

Commands may be added or changed frequently during development, and this list may not be up to date.

## Developing

The dryad repository is a dryad garden, so to develop you need a working dryad install.

You can either install a previous release, or run a "bootstrap" build running `make bootstrap-build` and `make bootstrap-install`.  You'll need go installed to build the bootstrap.

Once the bootstrap is installed, you can run `make dev-shell` to build and enter a dev shell.  This shell links to the dryad build in the garden, so your `dryad` executable inside the shell is updated every time you run `dryad garden build`.  If you break the dev shell environement, you can exit the dev shell, and use the bootstrap install to rebuild the dev shell until it works again.
