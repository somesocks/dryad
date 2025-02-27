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
		Include []string
		Exclude []string
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
				Include: includeOpts,
				Exclude: excludeOpts,
			}
		},
	)

	var listRoots = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
		if err != nil {
			return err, nil
		}

		err = roots.Walk(
			ctx,
			dryad.RootsWalkRequest{
				OnMatch: func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, any) {
					var err error
					var matchesInclude = false
					var matchesExclude = false

					if len(args.Include) == 0 { matchesInclude = true }

					for _, include := range args.Include {
						var matchesFilter bool
						err, matchesFilter = root.Filter(
							ctx,
							dryad.RootFilterRequest{
								Expression: include,
							},
						)
						if err != nil {
							return err, nil
						}
						matchesInclude = matchesInclude || matchesFilter
						if matchesInclude {
							break
						}
					}

					for _, exclude := range args.Exclude {
						var matchesFilter bool
						err, matchesFilter = root.Filter(
							ctx,
							dryad.RootFilterRequest{
								Expression: exclude,
							},
						)
						if err != nil {
							return err, nil
						}
						matchesExclude = matchesExclude || matchesFilter
						if matchesExclude {
							break
						}
					}

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(root.Roots.Garden.BasePath, root.BasePath)
					if err != nil {
						return err, nil
					}

					// zlog.Info().
					// 	Str("root", root.BasePath).
					// 	Bool("matchesFilter", matchesFilter).
					// 	Bool("args.Include(relPath)", args.Include(relPath)).
					// 	Bool("!args.Exclude(relPath)", !args.Exclude(relPath)).
					// 	Msg("roots list matchesFilter")

					if matchesInclude && !matchesExclude {
						if args.Relative {
							fmt.Println(relPath)
						} else {
							fmt.Println(root.BasePath)
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
		WithOption(clib.NewOption("include", "choose which roots are included in the list. the include filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the list.  the exclude filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
