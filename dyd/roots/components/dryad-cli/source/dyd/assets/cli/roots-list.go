package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootsListCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		Relative bool
		Parallel int
		IncludeRoots func(path string) bool
		ExcludeRoots func(path string) bool
	}	

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var relative bool = true
			var path string = ""

			if len(args) > 0 {
				path = args[0]
			}

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			var includeOpts []string
			var excludeOpts []string

			if options["exclude"] != nil {
				excludeOpts = options["exclude"].([]string)
			}

			if options["include"] != nil {
				includeOpts = options["include"].([]string)
			}

			includeRoots := dryad.RootIncludeMatcher(includeOpts)
			excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
	
			return nil, ParsedArgs{
				GardenPath: path,
				Parallel: parallel,
				Relative: relative,
				IncludeRoots: includeRoots,
				ExcludeRoots: excludeRoots,
			}
		},
	)
		
	var listRoots = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.UnsafeGardenReference{
			BasePath: args.GardenPath,
		}
		
		err, garden := unsafeGarden.Resolve(ctx, nil)
		if err != nil {
			return err, nil
		}

		err, _ = dryad.RootsWalk(
			ctx,
			dryad.RootsWalkRequest{
				Garden: &garden,
				OnRoot: func (ctx *task.ExecutionContext, match dryad.RootsWalkMatch) (error, any) {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(match.GardenPath, match.RootPath)
					if err != nil {
						return err, nil
					}


					if args.IncludeRoots(relPath) && !args.ExcludeRoots(relPath) {
						if args.Relative {
							fmt.Println(relPath)
						} else {
							fmt.Println(match.RootPath)
						}
					}

					return nil, nil
				},
			},
		)
		return err, nil
	}

	listRoots = task.WithContext(
		listRoots,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)


	var action = task.Return(
		task.Series2(
			parseArgs,
			listRoots,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating garden")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)").
		WithArg(
			clib.
				NewArg("path", "path to the base root (or garden) to list roots in").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("include", "choose which roots are included in the list").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the list").WithType(clib.OptionTypeMultiString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
