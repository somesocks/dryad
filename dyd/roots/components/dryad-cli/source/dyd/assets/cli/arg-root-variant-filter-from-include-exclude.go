package cli

import (
	"context"
	clib "dryad/cli-builder"
	"dryad/core"
)

var ArgRootVariantFilterFromIncludeExclude = func(
	ctx context.Context,
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

	return core.RootVariantSelectorFilter(
		core.SelectorFilterRequest{
			Include: includeOpts,
			Exclude: excludeOpts,
		},
	)
}
