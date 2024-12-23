package cli

import (
	clib "dryad/cli-builder"
)

var stemsCommand = clib.NewCommand("stems", "commands to work with dryad stems").
	WithCommand(stemsListCommand).
	WithCommand(stemsPathCommand)
