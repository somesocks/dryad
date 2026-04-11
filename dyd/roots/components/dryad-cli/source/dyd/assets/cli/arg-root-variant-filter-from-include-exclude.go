package cli

import (
	clib "dryad/cli-builder"
	"dryad/core"
	"dryad/task"
)

var ArgRootVariantFilterFromIncludeExclude = func(
	ctx *task.ExecutionContext,
	req clib.ActionRequest,
) (error, core.RootVariantFilter) {
	options := req.Opts

	var includeOpts []string
	var excludeOpts []string

	if options["exclude"] != nil {
		excludeOpts = options["exclude"].([]string)
	}

	if options["include"] != nil {
		includeOpts = options["include"].([]string)
	}

	return core.RootVariantCelFilter(
		core.RootCelFilterRequest{
			Include: includeOpts,
			Exclude: excludeOpts,
		},
	)
}
