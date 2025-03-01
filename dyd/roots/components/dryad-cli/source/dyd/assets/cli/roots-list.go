package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"path/filepath"

	"bufio"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootsListCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		Relative bool
		Parallel int
		FromStdinFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
		IncludeExcludeFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
	}


	var buildStdinFilter = func (
		ctx *task.ExecutionContext,
		req clib.ActionRequest,
	) (error, func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)) {
		var options = req.Opts

		var fromStdin bool
		var fromStdinFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)

		var path = ""

		if options["from-stdin"] != nil {
			fromStdin = options["from-stdin"].(bool)
		} else {
			fromStdin = false
		}

		if fromStdin {
			unsafeGarden := dryad.Garden(path)
	
			err, garden := unsafeGarden.Resolve(ctx)
			if err != nil {
				return err, fromStdinFilter
			}
	
			err, roots := garden.Roots().Resolve(ctx)
			if err != nil {
				return err, fromStdinFilter
			}

			var rootSet = make(map[string]bool)
			var scanner = bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				var path = scanner.Text()
				var err error 
				var root dryad.SafeRootReference

				path, err = filepath.Abs(path)
				if err != nil {
					zlog.Error().
						Err(err).
						Msg("error reading path from stdin")
					return err, fromStdinFilter
				}

				path = _rootsOwningDependencyCorrection(path)
				err, root = roots.Root(path).Resolve(ctx)
				if err != nil {
					zlog.Error().
						Str("path", path).
						Err(err).
						Msg("error resolving root from path")
					return err, fromStdinFilter
				}

				rootSet[root.BasePath] = true
			}

			// Check for any errors during scanning
			if err := scanner.Err(); err != nil {
				zlog.Error().Err(err).Msg("error reading stdin")
				return err, fromStdinFilter
			}

			fromStdinFilter = func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, bool) {
				_, ok := rootSet[root.BasePath]
				return nil, ok
			}

		} else {
			fromStdinFilter = func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, bool) {
				return nil, true
			}
		}

		return nil, fromStdinFilter
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

			if options["include"] != nil {
				includeOpts = options["include"].([]string)
			}

			if options["exclude"] != nil {
				excludeOpts = options["exclude"].([]string)
			}

			err, rootFilter := dryad.RootCelFilter(
				dryad.RootCelFilterRequest{
					Include: includeOpts,
					Exclude: excludeOpts,
				},
			)
			if err != nil {
				return err, ParsedArgs{}
			}


			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			err, fromStdinFilter := buildStdinFilter(task.SERIAL_CONTEXT, req)
			if err != nil {
				return err, ParsedArgs{}
			}
	
			return nil, ParsedArgs{
				GardenPath: path,
				Parallel: parallel,
				Relative: relative,
				FromStdinFilter: fromStdinFilter,
				IncludeExcludeFilter: rootFilter,
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
					var shouldMatch bool

					err, shouldMatch = args.FromStdinFilter(ctx, root)
					if err != nil {
						return err, nil
					} else if !shouldMatch {
						return nil, nil
					}

					err, shouldMatch = args.IncludeExcludeFilter(ctx, root)
					if err != nil {
						return err, nil
					} else if !shouldMatch {
						return nil, nil
					}

					if args.Relative {
						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(root.Roots.Garden.BasePath, root.BasePath)
						if err != nil {
							return err, nil
						}

						fmt.Println(relPath)
						return nil, nil
					} else {
						fmt.Println(root.BasePath)
						return nil, nil
					}


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
				NewArg("path", "path to the base garden to list roots in").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"relative", 
				"print roots relative to the base garden path. default true",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"from-stdin", 
				"if set, read a list of roots from stdin to use as a base list to print, instead of all roots. include and exclude filters will be applied to this list. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(clib.NewOption("include", "choose which roots are included in the list. the include filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the list.  the exclude filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
