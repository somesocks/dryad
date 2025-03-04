
package cli

import (
	clib "dryad/cli-builder"
	"dryad/core"
	"dryad/task"
	// "fmt"

	// zlog "github.com/rs/zerolog/log"
)

var ArgRootFilterFromIncludeExclude = func (
	ctx *task.ExecutionContext,
	req clib.ActionRequest,
) (error, core.RootFilter) {
	var options = req.Opts

	var includeOpts []string
	var excludeOpts []string

	if options["exclude"] != nil {
		excludeOpts = options["exclude"].([]string)
	}

	if options["include"] != nil {
		includeOpts = options["include"].([]string)
	}

	err, rootFilter := core.RootCelFilter(
		core.RootCelFilterRequest{
			Include: includeOpts,
			Exclude: excludeOpts,
		},
	)
	if err != nil {
		return err, nil
	}

	return nil, rootFilter
}
