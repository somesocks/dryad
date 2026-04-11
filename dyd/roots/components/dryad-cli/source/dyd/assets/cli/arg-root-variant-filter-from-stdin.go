package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	"dryad/core"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
)

var ArgRootVariantFilterFromStdin = func(
	ctx *task.ExecutionContext,
	req clib.ActionRequest,
) (error, core.RootVariantFilter) {
	options := req.Opts

	var fromStdin bool
	if options["from-stdin"] != nil {
		fromStdin = options["from-stdin"].(bool)
	}

	if !fromStdin {
		return nil, func(ctx *task.ExecutionContext, variant *core.SafeRootVariantReference) (error, bool) {
			return nil, true
		}
	}

	err, garden := core.Garden("").Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, roots := garden.Roots().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	variantSet := map[string]bool{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		raw := scanner.Text()

		err, rootRef := parseRootRef(raw)
		if err != nil {
			return err, nil
		}

		path, err := filepath.Abs(rootRef.Path)
		if err != nil {
			return err, nil
		}
		path = rootsInputOwnershipDependencyCorrection(path)

		err, root := roots.Root(path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, variants := root.ResolveBuildVariantReferences(
			ctx,
			core.RootResolveBuildVariantsRequest{
				Selector: rootRef.Selector,
			},
		)
		if err != nil {
			return err, nil
		}

		for _, variant := range variants {
			err, variantURL := variant.URL()
			if err != nil {
				return err, nil
			}
			variantSet[variant.Root.BasePath+variantURL] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return err, nil
	}

	return nil, func(ctx *task.ExecutionContext, variant *core.SafeRootVariantReference) (error, bool) {
		err, variantURL := variant.URL()
		if err != nil {
			return err, false
		}

		return nil, variantSet[variant.Root.BasePath+variantURL]
	}
}
