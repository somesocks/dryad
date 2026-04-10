package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"

	// "bufio"
	// "os"

	zlog "github.com/rs/zerolog/log"
)

var rootsListCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath           string
		Relative             bool
		ToSprouts            bool
		Parallel             int
		FromStdinFilter      dryad.RootVariantFilter
		IncludeExcludeFilter dryad.RootVariantFilter
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

		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, rootFilter := ArgRootVariantFilterFromIncludeExclude(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, fromStdinFilter := ArgRootVariantFilterFromStdin(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			GardenPath:           path,
			Parallel:             parallel,
			Relative:             relative,
			ToSprouts:            toSprouts,
			FromStdinFilter:      fromStdinFilter,
			IncludeExcludeFilter: rootFilter,
		}
	}

	var listRoots = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		err = roots.WalkVariants(
			ctx,
			dryad.RootsWalkVariantsRequest{
				ShouldMatch: dryad.RootVariantFiltersCompose(
					args.FromStdinFilter,
					args.IncludeExcludeFilter,
				),
				OnMatch: func(ctx *task.ExecutionContext, variant *dryad.SafeRootVariantReference) (error, any) {
					if args.ToSprouts {
						rootPath, err := filepath.Rel(variant.Root.Roots.BasePath, variant.Root.BasePath)
						if err != nil {
							return err, nil
						}

						var sproutPath string
						if args.Relative {
							sproutPath = filepath.Join("dyd", "sprouts", rootPath)
						} else {
							sproutPath = filepath.Join(sprouts.BasePath, rootPath)
						}

						err, variantURL := variant.URL()
						if err != nil {
							return err, nil
						}

						fmt.Println(sproutPath + variantURL)
					} else {
						err, variantRef := formatRootVariantRef(variant, args.Relative)
						if err != nil {
							return err, nil
						}
						fmt.Println(variantRef)
					}
					return nil, nil

				},
			},
		)
		return err, nil
	}

	listRoots = task.WithContext(
		listRoots,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listRoots,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while creating garden")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all root variants that are dependencies for the current root (or root variants of the current garden, if the path is not a root)").
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
				"if set, read a list of root refs from stdin to use as a base list to print, instead of all root variants. include and exclude filters will be applied to this list. default false",
			).
				WithType(clib.OptionTypeBool),
		).
		WithOption(clib.NewOption("include", "choose which root variants are included in the list. the include filter is a CEL expression with access to a 'root' object for each root variant.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which root variants are excluded from the list. the exclude filter is a CEL expression with access to a 'root' object for each root variant.").WithType(clib.OptionTypeMultiString)).
		WithOption(
			clib.NewOption(
				"to-sprouts",
				"if set, print the corresponding sprout ref for each root variant instead of the root ref.",
			).
				WithType(clib.OptionTypeBool),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
