package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
	"sort"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func rootVariantsListEncodeDescriptor(descriptor dryad.VariantDescriptor) string {
	if len(descriptor) == 0 {
		return "default"
	}

	keys := make([]string, 0, len(descriptor))
	for key := range descriptor {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+descriptor[key])
	}

	return strings.Join(parts, "+")
}

var rootVariantsListCommand = func() clib.Command {
	type ParsedArgs struct {
		RootPath string
		Parallel int
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts
		var rootPath string
		var err error

		if len(args) > 0 {
			rootPath = args[0]
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

		return nil, ParsedArgs{
			RootPath: rootPath,
			Parallel: parallel,
		}
	}

	var listVariants = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeRoot := roots.Root(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, variants := safeRoot.ResolveBuildVariants(ctx, dryad.RootResolveBuildVariantsRequest{})
		if err != nil {
			return err, nil
		}

		for _, variant := range variants {
			fmt.Println(rootVariantsListEncodeDescriptor(variant))
		}

		return nil, nil
	}

	listVariants = task.WithContext(
		listVariants,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listVariants,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing root variants")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all build variants of this root").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
