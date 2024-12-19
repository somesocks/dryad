package cli

import (
	clib "dryad/cli-builder"
)

var runCommand = func() clib.Command {
	command := clib.NewCommand("run", "alias for `dryad script run`").
		WithArg(
			clib.
				NewArg("command", "alias command").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the alias to exec").WithType(clib.OptionTypeBool)).
		WithArg(clib.NewArg("-- args", "args to pass to the command").AsOptional()).
		WithAction(scriptRunAction)

	command = LoggingCommand(command)


	return command
}()
