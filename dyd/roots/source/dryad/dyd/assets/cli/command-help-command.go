package cli

import (
	clib "dryad/cli-builder"
)

var HelpCommand = func(
	command clib.Command,
) clib.Command {
	return command.
		WithOption(clib.NewOption("help", "display help text for this command").WithType(clib.OptionTypeBool))
}
