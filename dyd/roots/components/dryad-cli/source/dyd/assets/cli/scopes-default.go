package cli

import (
	clib "dryad/cli-builder"
)

var scopesDefaultCommand = clib.NewCommand("default", "work with the default scope").
	WithCommand(scopesDefaultGetCommand).
	WithCommand(scopesDefaultSetCommand).
	WithCommand(scopesDefaultUnsetCommand)
