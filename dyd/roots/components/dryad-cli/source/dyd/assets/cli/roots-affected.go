package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"sort"

	zlog "github.com/rs/zerolog/log"
)

var rootsAffectedCommand = func() clib.Command {
	type ParsedArgs struct {
		Relative   bool
		Parallel   int
		GardenPath string
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var options = req.Opts

		var relative bool = true
		var parallel int
		var wd string
		var err error

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

		wd, err = os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			Relative:   relative,
			Parallel:   parallel,
			GardenPath: wd,
		}
	}

	var findAffectedRoots = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.GardenPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		changedPathsByRoot := make(map[string][]string)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			err, owningPath, changedPath := rootsInputOwnershipPaths(ctx, scanner.Text())
			if err != nil {
				return err, nil
			}
			err, root := roots.Root(owningPath).Resolve(ctx)
			if err == nil {
				changedPathsByRoot[root.BasePath] = append(changedPathsByRoot[root.BasePath], changedPath)
			}
		}

		// Check for any errors during scanning
		if err := scanner.Err(); err != nil {
			return err, nil
		}

		startNodes := make([]string, 0)
		startNodeSet := make(dryad.TStringSet)

		for rootPath, changedPaths := range changedPathsByRoot {
			err, root := roots.Root(rootPath).Resolve(ctx)
			if err != nil {
				return err, nil
			}

			err, affectedVariants := root.ResolveAffectedVariants(ctx, changedPaths)
			if err != nil {
				return err, nil
			}

			renderedRootPath := root.BasePath
			if args.Relative {
				renderedRootPath, err = filepath.Rel(garden.BasePath, renderedRootPath)
				if err != nil {
					return err, nil
				}
			}

			for _, affectedVariant := range affectedVariants {
				err, node := formatVariantDescriptorRef(renderedRootPath, affectedVariant)
				if err != nil {
					return err, nil
				}
				if startNodeSet[node] {
					continue
				}
				startNodeSet[node] = true
				startNodes = append(startNodes, node)
			}
		}

		err, graph := roots.Graph(
			ctx,
			dryad.RootsGraphRequest{
				Relative: args.Relative,
			},
		)
		if err != nil {
			return err, nil
		}

		graph = graph.Transpose()

		affectedNodes := graph.DescendantNodes(make(dryad.TStringSet), startNodes)
		for node := range startNodeSet {
			affectedNodes[node] = true
		}

		affectedList := affectedNodes.ToArray([]string{})
		sort.Strings(affectedList)

		for _, affectedNode := range affectedList {
			fmt.Println(affectedNode)
		}

		return nil, nil
	}

	findAffectedRoots = task.WithContext(
		findAffectedRoots,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			findAffectedRoots,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while finding affected roots")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("affected", "take a list of files from stdin, and print a list of root variants that may depend on those files").
		WithOption(clib.NewOption("relative", "print root refs relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
