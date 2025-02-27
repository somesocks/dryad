package cli

import (
	clib "dryad/cli-builder"
)

var rootsCommand = clib.NewCommand("roots", "commands to work with dryad roots").
	WithCommand(rootsAffectedCommand).
	WithCommand(rootsBuildCommand).
	WithCommand(rootsEachCommand).
	WithCommand(rootsGraphCommand).
	WithCommand(rootsListCommand).
	WithCommand(rootsOwningCommand).
	WithCommand(rootsPathCommand)
