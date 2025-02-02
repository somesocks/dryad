package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var gardenPathCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args

			var path string
			// var err error

			if len(args) > 0 {
				path = args[0]
			}

			return nil, ParsedArgs{
				GardenPath: path,
			}
		},
	)

	var printGardenPath = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		path, err := dryad.GardenPath(args.GardenPath)
		if err != nil {
			return err, nil
		}
		fmt.Println(path)
		return nil, nil
	}

	printGardenPath = task.WithContext(
		printGardenPath,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, ctx
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printGardenPath,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			return 0
		},
	)


	command := clib.NewCommand("path", "return the base path for a garden").
		WithArg(
			clib.
				NewArg("path", "the target path at which to start for the base garden path").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = LoggingCommand(command)


	return command
}()
