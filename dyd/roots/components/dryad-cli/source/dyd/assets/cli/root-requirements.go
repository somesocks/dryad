package cli

import (
	clib "dryad/cli-builder"
)

var rootRequirementsCommand = clib.
	NewCommand("requirements", "commands to work with all requirements of a root").
	WithCommand(rootRequirementsListCommand)
