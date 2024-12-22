package cli

import (
	clib "dryad/cli-builder"
)

var scriptsCommand = clib.NewCommand("scripts", "commands to work with scoped scripts").
	WithCommand(scriptsListCommand)
