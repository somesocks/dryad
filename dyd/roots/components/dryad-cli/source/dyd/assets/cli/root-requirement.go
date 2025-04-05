package cli

import (
	clib "dryad/cli-builder"
)

var rootRequirementCommand = clib.
	NewCommand("requirement", "commands to work with a single requirement of a root").
	WithCommand(rootRequirementAddCommand).
	WithCommand(rootRequirementRemoveCommand)
