package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a new root at the target path").
		WithArg(
			clib.
				NewArg("path", "the path to create the new root at").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string = args[0]

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: "",
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			unsafeRoot := dryad.UnsafeRootReference{
				BasePath: path,
				Garden: &garden,
			}
			unsafeRoot = unsafeRoot.Clean()


			err, safeRoot := dryad.RootCreate(
				task.SERIAL_CONTEXT,
				dryad.RootCreateRequest{
					Root: &unsafeRoot,
				},
			)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating root")
				return 1
			}

			fmt.Println(safeRoot.BasePath)
			return 0
		})

	command = LoggingCommand(command)


	return command
}()
