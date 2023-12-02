package cli

import (
	clib "dryad/cli-builder"
)

var rootsCommand = clib.NewCommand("roots", "commands to work with dryad roots").
	WithCommand(rootsBuildCommand).
	WithCommand(rootsListCommand).
	WithCommand(rootsOwningCommand).
	WithCommand(rootsPathCommand)
