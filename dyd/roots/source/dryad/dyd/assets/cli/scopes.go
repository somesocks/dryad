package cli

import (
	clib "dryad/cli-builder"
)

var scopesCommand = clib.NewCommand("scopes", "commands to work with scopes").
	WithCommand(scopesDefaultCommand).
	WithCommand(scopesListCommand).
	WithCommand(scopesPathCommand)
