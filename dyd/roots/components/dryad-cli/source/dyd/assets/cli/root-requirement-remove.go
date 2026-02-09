package cli

import (
	clib "dryad/cli-builder"
	core "dryad/core"
	task "dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementRemoveCommand = func() clib.Command {
	type ParsedArgs struct {
		RequirementName string
		RootPath        string
		Parallel        int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var requirementName = args[0]
			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			var rootPath, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				RequirementName: requirementName,
				RootPath:        rootPath,
				Parallel:        parallel,
			}
		},
	)

	var removeRequirement = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := core.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, root := roots.Root(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, requirements := root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		} else if requirements == nil {
			return fmt.Errorf("root has no requirements"), nil
		}

		err, requirement := requirements.Requirement(args.RequirementName).Resolve(ctx)
		if err != nil {
			return err, nil
		} else if requirement == nil {
			return fmt.Errorf("requirement does not exist"), nil
		}

		err = requirement.Remove(ctx)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	removeRequirement = task.WithContext(
		removeRequirement,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			removeRequirement,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while unlinking root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("remove", "remove a requirement from the current root").
		WithArg(
			clib.
				NewArg("name", "name of the requirement to remove").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
