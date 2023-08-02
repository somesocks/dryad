package cli

import (
	clib "dryad/cli-builder"
)

var scopeSettingCommand = clib.NewCommand("setting", "commands to work with scope settings").
	WithCommand(scopeSettingGetCommand).
	WithCommand(scopeSettingSetCommand).
	WithCommand(scopeSettingUnsetCommand)
