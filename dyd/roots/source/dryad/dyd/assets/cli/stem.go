package cli

import (
	clib "dryad/cli-builder"
)

var stemCommand = clib.NewCommand("stem", "commands to work with dryad stems").
	WithCommand(stemFingerprintCommand).
	WithCommand(stemFilesCommand).
	WithCommand(stemPackCommand).
	WithCommand(stemPathCommand).
	WithCommand(stemRunCommand).
	WithCommand(stemUnpackCommand)
