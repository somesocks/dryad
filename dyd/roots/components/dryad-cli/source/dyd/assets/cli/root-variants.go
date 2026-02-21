package cli

import (
	clib "dryad/cli-builder"
)

var rootVariantsCommand = clib.
	NewCommand("variants", "commands to work with all variants of a root").
	WithCommand(rootVariantsListCommand)
