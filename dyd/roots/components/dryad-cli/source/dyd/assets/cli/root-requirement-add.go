package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementAddCommand = func() clib.Command {
	type ParsedArgs struct {
		RootPath string
		DepPath  string
		Alias    string
		Parallel int
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var rootPath, err = os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		var depPath = args[0]
		var alias = ""
		if len(args) > 1 {
			alias = args[1]
		}

		var parallel int
		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, rootPath = dydfs.PartialEvalSymlinks(ctx, rootPath)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, depPath = dydfs.PartialEvalSymlinks(ctx, depPath)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			RootPath: rootPath,
			DepPath:  depPath,
			Alias:    alias,
			Parallel: parallel,
		}
	}

	var addRequirement = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
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

		err, reqs := root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, dep := roots.Root(args.DepPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = reqs.Add(
			ctx,
			dryad.RootRequirementsAddRequest{
				Dependency: &dep,
				Alias:      args.Alias,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	addRequirement = task.WithContext(
		addRequirement,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			addRequirement,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while linking root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("add", "add a root as a dependency of the current root").
		WithArg(
			clib.
				NewArg("path", "path to the root you want to add as a dependency").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(clib.NewArg("alias", "the alias to add the root under. if not specified, this defaults to the basename of the linked root").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
