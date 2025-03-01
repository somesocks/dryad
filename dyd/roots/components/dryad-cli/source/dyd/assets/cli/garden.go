package cli

import (
	clib "dryad/cli-builder"
)

var gardenCommand = clib.NewCommand("garden", "commands to work with a dryad garden").
	WithCommand(gardenCreateCommand).
	WithCommand(gardenPackCommand).
	WithCommand(gardenPathCommand).
	WithCommand(gardenPruneCommand).
	WithCommand(gardenWipeCommand)
