package cli

import (
	clib "dryad/cli-builder"
)

var developCommand = clib.NewCommand("develop", "alias for root develop").
	WithCommand(rootDevelopStartCommand).
	WithCommand(rootDevelopStatusCommand).
	WithCommand(rootDevelopSaveCommand)
