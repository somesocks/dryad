package cli

import (
	clib "dryad/cli-builder"
)

func BuildCLI(
	Version string,
	Fingerprint string,
) clib.App {

	var app = clib.New("dryad package manager " + Version)

	app = app.
		WithCommand(gardenCommand).
		WithCommand(rootCommand).
		WithCommand(rootsCommand).
		WithCommand(runCommand).
		WithCommand(scopeCommand).
		WithCommand(scopesCommand).
		WithCommand(scriptCommand).
		WithCommand(scriptsCommand).
		WithCommand(sproutCommand).
		WithCommand(sproutsCommand).
		WithCommand(stemCommand).
		WithCommand(stemsCommand).
		WithCommand(systemCommand).
		WithCommand(versionCommand(Version, Fingerprint)).
		WithOption(clib.NewOption("help", "display help text for this command").WithType(clib.OptionTypeBool))

	return app
}
