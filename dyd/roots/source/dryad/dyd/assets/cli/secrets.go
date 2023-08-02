package cli

import (
	clib "dryad/cli-builder"
)

var secretsCommand = clib.NewCommand("secrets", "commands to work with dryad secrets").
	WithCommand(secretsFingerprintCommand).
	WithCommand(secretsListCommand).
	WithCommand(secretsPathCommand)
