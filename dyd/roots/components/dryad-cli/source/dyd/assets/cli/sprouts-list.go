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
		IncludeSprouts func(path string) bool
		ExcludeSprouts func(path string) bool
	}	

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var relative bool = true
			var path string = ""
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			var gardenPath string
			gardenPath, err = dryad.GardenPath(path)
			if err != nil {
				return err, ParsedArgs{}
			}

			var includeOpts []string
			var excludeOpts []string

			if options["exclude"] != nil {
				excludeOpts = options["exclude"].([]string)
			}

			if options["include"] != nil {
				includeOpts = options["include"].([]string)
			}

			includeSprouts := dryad.RootIncludeMatcher(includeOpts)
			excludeSprouts := dryad.RootExcludeMatcher(excludeOpts)

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
	
			return nil, ParsedArgs{
				GardenPath: gardenPath,
				Parallel: parallel,
				Relative: relative,
				IncludeSprouts: includeSprouts,
				ExcludeSprouts: excludeSprouts,
			}
		},
	)
		
	var listSprouts = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, _ := dryad.SproutsWalk(
			ctx,
			dryad.SproutsWalkRequest{
				GardenPath: args.GardenPath,
				OnSprout: func (ctx *task.ExecutionContext, path string) (error, any) {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(args.GardenPath, path)
					if err != nil {
						return err, nil
					}


					if args.IncludeSprouts(relPath) && !args.ExcludeSprouts(relPath) {
						if args.Relative {
							fmt.Println(relPath)
						} else {
							fmt.Println(path)
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
		WithOption(clib.NewOption("include", "choose which sprouts are included in the list").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded from the list").WithType(clib.OptionTypeMultiString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()
