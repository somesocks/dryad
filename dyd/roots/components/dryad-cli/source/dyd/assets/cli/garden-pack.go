package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var gardenPackCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		TargetPath string
		IncludeRoots bool
		IncludeHeap bool
		IncludeContexts bool
		IncludeSprouts bool
		IncludeShed bool
		Parallel int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var gardenPath = ""
			var targetPath = ""
			switch len(args) {
			case 0:
				break
			case 1:
				gardenPath = args[0]
			default:
				gardenPath = args[0]
				targetPath = args[1]
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			var includeRoots bool
			if options["include-roots"] != nil {
				includeRoots = options["include-roots"].(bool)
			} else {
				includeRoots = true
			}

			var includeHeap bool
			if options["include-heap"] != nil {
				includeHeap = options["include-heap"].(bool)
			} else {
				includeHeap = true
			}

			var includeContexts bool
			if options["include-contexts"] != nil {
				includeContexts = options["include-contexts"].(bool)
			} else {
				includeContexts = false
			}

			var includeSprouts bool
			if options["include-sprouts"] != nil {
				includeSprouts = options["include-sprouts"].(bool)
			} else {
				includeSprouts = true
			}

			var includeShed bool
			if options["include-shed"] != nil {
				includeShed = options["include-shed"].(bool)
			} else {
				includeShed = true
			}

			return nil, ParsedArgs{
				GardenPath: gardenPath,
				TargetPath: targetPath,
				IncludeRoots: includeRoots,
				IncludeHeap: includeHeap,
				IncludeContexts: includeContexts,
				IncludeSprouts: includeSprouts,
				IncludeShed: includeShed,
				Parallel: parallel,
			}
		},
	)

	var packGarden = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.UnsafeGardenReference{
			BasePath: args.GardenPath,
		}
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, targetPath := dryad.GardenPack(
			ctx, 
			dryad.GardenPackRequest{
				Garden: garden,
				TargetPath: args.TargetPath,
				IncludeRoots: args.IncludeRoots,
				IncludeHeap: args.IncludeHeap,
				IncludeContexts: args.IncludeContexts,
				IncludeSprouts: args.IncludeSprouts,
				IncludeShed: args.IncludeShed,
			},
		)
		fmt.Println(targetPath)
		return err, nil
	}

	packGarden = task.WithContext(
		packGarden,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			packGarden,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while packing garden")
				return 1
			}

			return 0
		},
	)


	command := clib.NewCommand("pack", "pack the current garden into an archive").
		WithOption(
			clib.
				NewOption("include-roots", "include dyd/roots in the archive. default true").
				WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.
				NewOption("include-heap", "include dyd/heap in the archive. default true").
				WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.
				NewOption("include-contexts", "include dyd/heap/contexts in the archive. default false").
				WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.
				NewOption("include-sprouts", "include dyd/sprouts in the archive. default true").
				WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.
				NewOption("include-shed", "include dyd/shed in the archive. default true").
				WithType(clib.OptionTypeBool),
		).
		WithArg(
			clib.
				NewArg("gardenPath", "the path to the garden to pack").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("targetPath", "the path (including name) to output the archive to").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
