package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	fs2 "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"

	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var stemAncestorsCommand = func() clib.Command {

	type ParsedArgs struct {
		Path     string
		Relative bool
		Self     bool
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var err error
		var path string
		var relative bool = true
		var self bool = false
		var parallel int

		if len(args) > 0 {
			path = args[0]
			path, err = filepath.Abs(path)
			if err != nil {
				return err, ParsedArgs{}
			}
		}

		if path == "" {
			path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}
		}

		if options["relative"] != nil {
			relative = options["relative"].(bool)
		}

		if options["self"] != nil {
			self = options["self"].(bool)
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Path:     path,
			Relative: relative,
			Self:     self,
			Parallel: parallel,
		}
	}

	var runAncestors = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.Path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.StemAncestorsWalk(
			dryad.StemAncestorsWalkRequest{
				BasePath: args.Path,
				OnMatch: func(node fs2.Walk6Node) error {
					relPath, err := filepath.Rel(garden.BasePath, node.Path)
					if err != nil {
						return err
					}

					if args.Relative {
						fmt.Println(relPath)
					} else {
						fmt.Println(node.Path)
					}

					return nil
				},
				Self: args.Self,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	runAncestors = task.WithContext(
		runAncestors,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runAncestors,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing stem ancestors")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("ancestors", "list all direct and indirect dependencies of a stem").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print stems relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("self", "include the base stem itself. default false").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
