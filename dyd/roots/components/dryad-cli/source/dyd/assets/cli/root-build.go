package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath string
		Parallel int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var path string

			var parallel int

			if len(args) > 0 {
				path = args[0]
			}

			if !filepath.IsAbs(path) {
				wd, err := os.Getwd()
				if err != nil {
					return err, ParsedArgs{}
				}
				path = filepath.Join(wd, path)
			}

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
				
			return nil, ParsedArgs{
				RootPath: path,
				Parallel: parallel,
			}
		},
	)

	var buildRoot = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		var rootFingerprint string
		err, rootFingerprint := dryad.RootBuild(
			task.SERIAL_CONTEXT,
			dryad.RootBuildRequest{
				RootPath: args.RootPath,	
			},
		)
		if err != nil {
			return err, nil
		}
		fmt.Println(rootFingerprint)

		return nil, nil
	}

	buildRoot = task.WithContext(
		buildRoot,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			buildRoot,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building root")
				return 1
			}

			return 0
		},
	)

	command := clib.
		NewCommand("build", "build a specified root").
		WithArg(
			clib.
				NewArg("path", "path to the root to build").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
