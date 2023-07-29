package cli

import (
	clib "dryad/cli-builder"
)

var scriptCommand = clib.NewCommand("script", "commands to work with a scoped script").
	WithCommand(scriptEditCommand).
	WithCommand(scriptGetCommand).
	WithCommand(scriptPathCommand).
	WithCommand(scriptRunCommand)
