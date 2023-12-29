package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultGetCommand = func() clib.Command {
	command := clib.NewCommand("get", "return the name of the default scope, if set").
		WithAction(func(req clib.ActionRequest) int {

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			scopeName, err := dryad.ScopeGetDefault(path)
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			if scopeName != "" {
				fmt.Println(scopeName)
			}

			return 0
		})
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
