package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var sproutsListCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		Relative bool
		Parallel int		
		Filter func (*task.ExecutionContext, *dryad.SafeSproutReference) (error, bool)
	}	

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var err error
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

			
			err, sproutFilter := dryad.SproutFilterFromCel(
				dryad.SproutFilterFromCelRequest{
					Include: includeOpts,
					Exclude: excludeOpts,
				},
			)
			if err != nil {
				return err, ParsedArgs{}
			}


			err, fromStdinFilter := ArgSproutFilterFromStdin(task.SERIAL_CONTEXT, req)
			if err != nil {
				return err, ParsedArgs{}
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}
	
			return nil, ParsedArgs{
				GardenPath: path,
				Parallel: parallel,
				Relative: relative,
				Filter: dryad.SproutFiltersCompose(
					fromStdinFilter,
					sproutFilter,
				),
			}
		},
	)
		
	var listSprouts = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = sprouts.Walk(
			ctx,
			dryad.SproutsWalkRequest{
				OnSprout: func (ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference) (error, any) {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(sprout.Sprouts.Garden.BasePath, sprout.BasePath)
					if err != nil {
						return err, nil
					}

					err, shouldMatch := args.Filter(ctx, sprout)
					if err != nil {
						return err, nil
					}

					if shouldMatch {
						if args.Relative {
							fmt.Println(relPath)
						} else {
							fmt.Println(sprout.BasePath)
						}
					}

					return nil, nil
				},
			},
		)
		return err, nil
	}

	listSprouts = task.WithContext(
		listSprouts,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)


	var action = task.Return(
		task.Series2(
			parseArgs,
			listSprouts,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error during sprouts list")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all sprouts of the current garden").
		WithOption(clib.NewOption("relative", "print sprouts relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(
			clib.NewOption(
				"from-stdin", 
				"if set, read a list of sprouts from stdin to use as a base list to print, instead of all sprouts. include and exclude filters are applied to this list. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(clib.NewOption("include", "choose which sprouts are included in the list. the include filter is a CEL expression with access to a 'sprout' object that can be used to filter on properties of each sprout.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded from the list.  the exclude filter is a CEL expression with access to a 'sprout' object that can be used to filter on properties of each sprout.").WithType(clib.OptionTypeMultiString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()
