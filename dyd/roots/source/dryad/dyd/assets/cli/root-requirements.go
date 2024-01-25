package cli

import (
	clib "dryad/cli-builder"
)

var rootRequirementsCommand = clib.NewCommand("requirements", "commands to work with the requirements of a root").
	WithCommand(rootRequirementsAddCommand).
	WithCommand(rootRequirementsListCommand).
	WithCommand(rootRequirementsRemoveCommand)
