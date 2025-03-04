package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"path/filepath"

	// "bufio"
	// "os"

	zlog "github.com/rs/zerolog/log"
)

var rootsListCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		Relative bool
		ToSprouts bool
		Parallel int
		FromStdinFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
		IncludeExcludeFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var relative bool = true
		var toSprouts bool
		var path string = ""

		if len(args) > 0 {
			path = args[0]
		}

		if options["relative"] != nil {
			relative = options["relative"].(bool)
		} else {
			relative = true
		}

		if options["to-sprouts"] != nil {
			toSprouts = options["to-sprouts"].(bool)
		} else {
			toSprouts = false
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
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, fromStdinFilter := ArgRootFilterFromStdin(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			GardenPath: path,
			Parallel: parallel,
			Relative: relative,
			ToSprouts: toSprouts,
			FromStdinFilter: fromStdinFilter,
			IncludeExcludeFilter: rootFilter,
		}
	}

	var listRoots = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = roots.Walk(
			ctx,
			dryad.RootsWalkRequest{
				ShouldMatch: dryad.RootFiltersCompose(
					args.FromStdinFilter,
					args.IncludeExcludeFilter,
				),
				OnMatch: func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, any) {
					if args.ToSprouts {
						// calculate the relative path to the root from the base of the roots
						rootPath, err := filepath.Rel(root.Roots.BasePath, root.BasePath)
						if err != nil {
							return err, nil
						}

						var sproutPath string
						if args.Relative {
							sproutPath = filepath.Join("dyd", "sprouts", rootPath)
						} else {
							sproutPath = filepath.Join(sprouts.BasePath, rootPath)
						}

						fmt.Println(sproutPath)
					} else if args.Relative {
						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(root.Roots.Garden.BasePath, root.BasePath)
						if err != nil {
							return err, nil
						}

						fmt.Println(relPath)
					} else {
						fmt.Println(root.BasePath)
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
		WithOption(
			clib.NewOption(
				"to-sprouts", 
				"if set, print the corresponding sprout path for each root instead of the root path.",
			).
			WithType(clib.OptionTypeBool),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
