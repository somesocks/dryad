## Development

The dryad repository is a dryad garden, so to develop you need a working dryad install.

You can either install a previous release, or run a "bootstrap" build running `make bootstrap-build` and `make bootstrap-install`.  You'll need go installed to build the bootstrap.

Once the bootstrap is installed, you can run `make dev-shell` to build and enter a dev shell.  This shell links to the dryad build in the garden, so your `dryad` executable inside the shell is updated every time you run `dryad garden build`.  If you break the dev shell environement, you can exit the dev shell, and use the bootstrap install to rebuild the dev shell until it works again.
