package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all stems that are dependencies for the current root").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding directory")
				return 1
			}
			err = dryad.StemsWalk(path, func(path string, info fs.FileInfo, err error) error {
				fmt.Println(path)
				return nil
			})
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling stems")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
