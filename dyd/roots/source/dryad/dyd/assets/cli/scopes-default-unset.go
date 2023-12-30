package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultUnsetCommand = func() clib.Command {
	command := clib.NewCommand("unset", "remove the default scope setting").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding active directory")
				return 1
			}

			err = dryad.ScopeUnsetDefault(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while removing active scope")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
