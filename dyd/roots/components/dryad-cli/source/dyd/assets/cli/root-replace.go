package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var rootReplaceCommand = func() clib.Command {
	type ParsedArgs struct {
		SourcePath           string
		DestPath             string
		Parallel             int
		FromStdinFilter      func(*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
		IncludeExcludeFilter func(*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var source string = args[0]
		var dest string = args[1]
		var err error
		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, includeExcludeFilter := ArgRootFilterFromIncludeExclude(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, fromStdinFilter := ArgRootFilterFromStdin(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, source = dydfs.PartialEvalSymlinks(ctx, source)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, dest = dydfs.PartialEvalSymlinks(ctx, dest)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			SourcePath:           source,
			DestPath:             dest,
			Parallel:             parallel,
			FromStdinFilter:      fromStdinFilter,
			IncludeExcludeFilter: includeExcludeFilter,
		}
	}

	var replaceRoot = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeSourceRoot := roots.Root(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeDestRoot := roots.Root(args.DestPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = safeSourceRoot.Replace(
			ctx,
			dryad.RootReplaceRequest{
				Filter: dryad.RootFiltersCompose(
					args.FromStdinFilter,
					args.IncludeExcludeFilter,
				),
				Dest: &safeDestRoot,
			},
		)

		return err, nil
	}

	replaceRoot = task.WithContext(
		replaceRoot,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			replaceRoot,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while replacing root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("replace", "replace all references to one root with references to another").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("replacement", "path to the replacement root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"include",
				"choose which roots are included in the search to find references to replace. the include filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithOption(
			clib.NewOption(
				"exclude",
				"choose which roots are excluded in the search to find references to replace.  the exclude filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithOption(
			clib.NewOption(
				"from-stdin",
				"if set, read a list of roots from stdin to use as a base list, instead of all roots. include and exclude filters will be applied to this list. default false",
			).
				WithType(clib.OptionTypeBool),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
