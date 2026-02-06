package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"

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

		rootSet := make(dryad.TStringSet)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			path := scanner.Text()
			path, err = filepath.Abs(path)
			if err != nil {
				return err, nil
			}
			path = _rootsOwningDependencyCorrection(path)
			err, root := roots.Root(path).Resolve(ctx)
			if err == nil {
				rootSet[root.BasePath] = true
			}
		}

		// Check for any errors during scanning
		if err := scanner.Err(); err != nil {
			return err, nil
		}

		rootList := rootSet.ToArray([]string{})

		err, graph := roots.Graph(
			ctx,
			dryad.RootsGraphRequest{
				Relative: false,
			},
		)
		if err != nil {
			return err, nil
		}

		graph = graph.Transpose()

		// find the descendants of the affected roots
		descendants := graph.Descendants(make(dryad.TStringSet), rootList)
		for k := range descendants {
			rootSet[k] = true
		}

		// Print the resulting roots
		if args.Relative {
			for key := range rootSet {
				// calculate the relative path to the root from the base of the garden
				relPath, err := filepath.Rel(garden.BasePath, key)
				if err != nil {
					return err, nil
				}
				fmt.Println(relPath)
			}
		} else {
			for key := range rootSet {
				fmt.Println(key)
			}
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
				zlog.Fatal().Err(err).Msg("error while finding affected roots")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("affected", "take a list of files from stdin, and print a list of roots that may depend on those files").
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
