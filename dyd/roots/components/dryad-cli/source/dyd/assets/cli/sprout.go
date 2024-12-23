package cli

import (
	clib "dryad/cli-builder"
)

var sproutCommand = clib.NewCommand("sprout", "commands to work with a single dryad sprout").
	WithCommand(sproutRunCommand)
