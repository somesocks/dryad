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

var _rootsOwningDependencyCorrection = func(path string) string {
	p1, _ := filepath.Split(path)
	p1 = filepath.Clean(p1)
	p2, f2 := filepath.Split(p1)
	p2 = filepath.Clean(p2)
	p3, f3 := filepath.Split(p2)
	p3 = filepath.Clean(p3)

	if f3 == "dyd" && f2 == "requirements" {
		return p3
	} else {
		return path
	}

}

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

		rootSet := make(map[string]bool)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			path := scanner.Text()
			path, err := filepath.Abs(path)
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
				zlog.Fatal().Err(err).Msg("error while finding owning roots")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("owning", "list all roots that are owners of the provided files. The files to check should be provided as relative or absolute paths through stdin.").
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
