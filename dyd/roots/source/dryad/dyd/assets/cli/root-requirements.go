package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsCommand = func() clib.Command {
	command := clib.NewCommand("requirements", "list all requirements of this root").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var root string

			if len(args) > 0 {
				root = args[0]
			}

			err := dryad.RootRequirementsWalk(root, func(path string, info fs.FileInfo) error {
				fmt.Println(path)
				return nil
			})

			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
