package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"sort"

	zlog "github.com/rs/zerolog/log"
)

var rootsOwningCommand = func() clib.Command {
	type ParsedArgs struct {
		Relative bool
		Parallel int
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var options = req.Opts

		var relative bool = true
		var parallel int

		if options["relative"] != nil {
			relative = options["relative"].(bool)
		} else {
			relative = true
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Relative: relative,
			Parallel: parallel,
		}
	}

	var findOwningRoots = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden("").Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		changedPathsByRoot := make(map[string][]string)
		ownerSet := make(dryad.TStringSet)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			rawPath := scanner.Text()
			err, rawChangedPath := rootsFileRequirementOwnershipPath(ctx, rawPath)
			if err != nil {
				return err, nil
			}
			err, owningPath, changedPath := rootsInputOwnershipPaths(ctx, rawPath)
			if err != nil {
				return err, nil
			}
			err, root := roots.Root(owningPath).Resolve(ctx)
			if err == nil {
				changedPathsByRoot[root.BasePath] = append(changedPathsByRoot[root.BasePath], changedPath)
			}

			err, fileOwners := roots.FileRequirementOwners(ctx, rawChangedPath)
			if err != nil {
				return err, nil
			}
			for _, fileOwner := range fileOwners {
				root := dryad.SafeRootReference{BasePath: fileOwner.RootPath, Roots: roots}
				err, ownerRef := formatRootVariantDescriptorRef(&root, fileOwner.Variant, args.Relative)
				if err != nil {
					return err, nil
				}
				ownerSet[ownerRef] = true
			}
		}

		// Check for any errors during scanning
		if err := scanner.Err(); err != nil {
			return err, nil
		}

		for rootPath, changedPaths := range changedPathsByRoot {
			err, root := roots.Root(rootPath).Resolve(ctx)
			if err != nil {
				return err, nil
			}

			err, owningVariants := root.ResolveAffectedVariants(ctx, changedPaths)
			if err != nil {
				return err, nil
			}

			for _, owningVariant := range owningVariants {
				err, ownerRef := formatRootVariantDescriptorRef(&root, owningVariant, args.Relative)
				if err != nil {
					return err, nil
				}
				ownerSet[ownerRef] = true
			}
		}

		owners := ownerSet.ToArray([]string{})
		sort.Strings(owners)

		for _, owner := range owners {
			fmt.Println(owner)
		}

		return nil, nil
	}

	findOwningRoots = task.WithContext(
		findOwningRoots,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			findOwningRoots,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while finding owning roots")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("owning", "list all root variants that own the provided paths. The paths to check should be provided as relative or absolute paths through stdin.").
		WithOption(clib.NewOption("relative", "print root refs relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
