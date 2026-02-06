package cli

import (
	clib "dryad/cli-builder"
)

var scriptCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create and edit a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("editor", "set the editor to use")).
		WithAction(scriptEditAction)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
