package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"
	"sort"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func rootRequirementsList_encodeVariantSelectorURL(selector dryad.VariantDescriptor) string {
	if len(selector) == 0 {
		return ""
	}

	keys := make([]string, 0, len(selector))
	for key := range selector {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+selector[key])
	}

	return "?" + strings.Join(parts, "&")
}

var rootRequirementsListCommand = func() clib.Command {
	type ParsedArgs struct {
		RootPath string
		Relative bool
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

		var relative bool

		if options["relative"] != nil {
			relative = options["relative"].(bool)
		} else {
			relative = true
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
			Relative: relative,
			Parallel: parallel,
		}
	}

	var listRequirements = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		err, safeRequirements := safeRoot.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		} else if safeRequirements == nil {
			// no requirements, so exit
			return nil, nil
		}

		var onRequirementMatch = func(ctx *task.ExecutionContext, requirement *dryad.SafeRootRequirementReference) (error, any) {
			zlog.Trace().
				Str("path", requirement.BasePath).
				Msg("root requirements list / onRequirement")

			err, targetSpec := requirement.TargetSpec(ctx)
			if err != nil {
				return err, nil
			}

			targetPath, err := filepath.Rel(
				filepath.Dir(requirement.BasePath),
				targetSpec.Root.BasePath,
			)
			if err != nil {
				return err, nil
			}
			targetURL := "root:" + targetPath + rootRequirementsList_encodeVariantSelectorURL(targetSpec.VariantSelector)

			requirementPath := requirement.BasePath
			if args.Relative {
				// calculate the relative path to the root from the base of the garden
				requirementPath, err = filepath.Rel(
					requirement.Requirements.Root.Roots.Garden.BasePath,
					requirement.BasePath,
				)
				if err != nil {
					return err, nil
				}
			}

			fmt.Println(requirementPath + " -> " + targetURL)
			return nil, nil
		}

		err = safeRequirements.Walk(
			ctx,
			dryad.RootRequirementsWalkRequest{
				OnMatch: onRequirementMatch,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	listRequirements = task.WithContext(
		listRequirements,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listRequirements,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while crawling root requirements")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all requirements of this root").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
