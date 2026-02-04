package cli

import (
	clib "dryad/cli-builder"
)

var rootDevelopCommand = clib.NewCommand("develop", "commands to work with root development environments").
	WithCommand(rootDevelopStartCommand).
	WithCommand(rootDevelopStatusCommand).
	WithCommand(rootDevelopSaveCommand).
	WithCommand(rootDevelopSnapshotCommand).
	WithCommand(rootDevelopResetCommand).
	WithCommand(rootDevelopStopCommand)
