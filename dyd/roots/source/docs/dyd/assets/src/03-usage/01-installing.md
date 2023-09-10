# Installing
Binary releases are provided on the [Github Releases page](https://github.com/somesocks/dryad/releases).

dryad is distributed as a single binary, so you only need to download the right version for your system and place in the environment path.

The recommended way to install dryad is as a user-local tool in a directory like `$HOME/bin` (one standard location for user-local binaries), and then make sure that directory is added to your path (by adding something like `export PATH="$HOME/bin":$PATH` to your shell profile or rc).

An install/update script is provided together with the release assets for each version that downloads and installs dryad in `$HOME/bin`. You can install dryad directly by running this script using a command like `curl -L -f <INSTALL_SCRIPT_URL> | sh`.  The install script does not add the directory to your path.

## Bash Auto-completion
For shells that support Bash-style tab auto-completion, a helper script is provided in releases to enable auto-completion for dryad commands and arguments. To enable it, download the bash auto-completion script and place it in your auto-completion config directory (for example, `/etc/bash_completion.d/`).  Auto-completion scripts are generally loaded on shell start, so you may need to close and re-open your shell.
