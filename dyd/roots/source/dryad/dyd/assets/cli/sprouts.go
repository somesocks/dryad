package cli

import (
	clib "dryad/cli-builder"
)

var sproutsCommand = clib.NewCommand("sprouts", "commands to work with dryad sprouts").
	WithCommand(sproutsListCommand).
	WithCommand(sproutsPathCommand).
	WithCommand(sproutsPruneCommand).
	WithCommand(sproutsRunCommand)
