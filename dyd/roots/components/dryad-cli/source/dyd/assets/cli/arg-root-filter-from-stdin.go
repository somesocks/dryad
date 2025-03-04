
package cli

import (
	clib "dryad/cli-builder"
	"dryad/core"
	"dryad/task"
	// "fmt"
	"path/filepath"

	"bufio"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var ArgRootFilterFromStdin = func (
	ctx *task.ExecutionContext,
	req clib.ActionRequest,
) (error, core.RootFilter) {
	var options = req.Opts

	var fromStdin bool
	var fromStdinFilter core.RootFilter

	var path = ""

	if options["from-stdin"] != nil {
		fromStdin = options["from-stdin"].(bool)
	} else {
		fromStdin = false
	}

	if fromStdin {
		unsafeGarden := core.Garden(path)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, fromStdinFilter
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, fromStdinFilter
		}

		var rootSet = make(map[string]bool)
		var scanner = bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			var path = scanner.Text()
			var err error 
			var root core.SafeRootReference

			path, err = filepath.Abs(path)
			if err != nil {
				zlog.Error().
					Err(err).
					Msg("error reading path from stdin")
				return err, fromStdinFilter
			}

			path = _rootsOwningDependencyCorrection(path)
			err, root = roots.Root(path).Resolve(ctx)
			if err != nil {
				zlog.Error().
					Str("path", path).
					Err(err).
					Msg("error resolving root from path")
				return err, fromStdinFilter
			}

			rootSet[root.BasePath] = true
		}

		// Check for any errors during scanning
		if err := scanner.Err(); err != nil {
			zlog.Error().Err(err).Msg("error reading stdin")
			return err, fromStdinFilter
		}

		fromStdinFilter = func (ctx *task.ExecutionContext, root *core.SafeRootReference) (error, bool) {
			_, ok := rootSet[root.BasePath]
			return nil, ok
		}

	} else {
		fromStdinFilter = func (ctx *task.ExecutionContext, root *core.SafeRootReference) (error, bool) {
			return nil, true
		}
	}

	return nil, fromStdinFilter
}
