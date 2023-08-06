package cli

import (
	clib "dryad/cli-builder"
)

var systemCommand = clib.NewCommand("system", "maintenance and utility commands for dryad").
	WithCommand(systemAutocomplete).
	WithCommand(systemCommands)
