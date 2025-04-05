package cli

import (
	clib "dryad/cli-builder"
)

var rootCommand = clib.NewCommand("root", "commands to work with a dryad root").
	WithCommand(rootAncestorsCommand).
	WithCommand(rootBuildCommand).
	WithCommand(rootCopyCommand).
	WithCommand(rootCreateCommand).
	WithCommand(rootDescendantsCommand).
	WithCommand(rootDevelopCommand).
	WithCommand(rootMoveCommand).
	WithCommand(rootPathCommand).
	WithCommand(rootReplaceCommand).
	WithCommand(rootRequirementCommand).
	WithCommand(rootRequirementsCommand).
	WithCommand(rootSecretsCommand)
