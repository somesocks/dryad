package cli

import (
	clib "dryad/cli-builder"
)

var rootSecretsCommand = clib.NewCommand("secrets", "commands to work with dryad secrets").
	WithCommand(rootSecretsFingerprintCommand).
	WithCommand(rootSecretsListCommand).
	WithCommand(rootSecretsPathCommand)
