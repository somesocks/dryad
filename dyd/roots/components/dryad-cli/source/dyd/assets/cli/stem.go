package cli

import (
	clib "dryad/cli-builder"
)

var stemCommand = clib.NewCommand("stem", "commands to work with a single dryad stem").
	WithCommand(stemAncestorsCommand).
	WithCommand(stemFingerprintCommand).
	WithCommand(stemFilesCommand).
	WithCommand(stemPackCommand).
	WithCommand(stemPathCommand).
	WithCommand(stemRunCommand).
	WithCommand(stemUnpackCommand)
