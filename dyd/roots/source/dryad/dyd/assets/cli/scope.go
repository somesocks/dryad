package cli

import (
	clib "dryad/cli-builder"
)

var scopeCommand = clib.NewCommand("scope", "commands to work with a single scope").
	WithCommand(scopeActiveCommand).
	WithCommand(scopeCreateCommand).
	WithCommand(scopeDeleteCommand).
	WithCommand(scopeUseCommand).
	WithCommand(scopeSettingCommand)
