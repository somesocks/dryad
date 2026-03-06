package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"net/url"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func rootRequirementAdd_parseDependencyTarget(raw string) (error, string, string) {
	targetURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err, "", ""
	}

	if targetURL.Scheme != "" && targetURL.Scheme != "root" {
		return fmt.Errorf("unsupported scheme for root requirement: %s", targetURL.Scheme), "", ""
	}
	if targetURL.Fragment != "" {
		return fmt.Errorf("variant descriptor fragments are not supported; use query parameters with '&'"), "", ""
	}

	targetPath := targetURL.Path
	if targetURL.Scheme == "root" && targetURL.Opaque != "" {
		targetPath = targetURL.Opaque
	}
	if targetPath == "" {
		return fmt.Errorf("missing root requirement target path"), "", ""
	}

	variantSelectorRaw := ""
	if targetURL.RawQuery != "" {
		variantSelectorRaw = "?" + targetURL.RawQuery
	}

	return nil, targetPath, variantSelectorRaw
}

var rootRequirementAddCommand = func() clib.Command {
	type ParsedArgs struct {
		RootPath           string
		DepPath            string
		DepVariantSelector string
		Alias              string
		Parallel           int
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var rootPath, err = os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		var depRaw = args[0]
		var alias = ""
		if len(args) > 1 {
			alias = args[1]
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

		err, depPath, depVariantSelector := rootRequirementAdd_parseDependencyTarget(depRaw)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, depPath = dydfs.PartialEvalSymlinks(ctx, depPath)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			RootPath:           rootPath,
			DepPath:            depPath,
			DepVariantSelector: depVariantSelector,
			Alias:              alias,
			Parallel:           parallel,
		}
	}

	var addRequirement = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, root := roots.Root(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, reqs := root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, dep := roots.Root(args.DepPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = reqs.Add(
			ctx,
			dryad.RootRequirementsAddRequest{
				Dependency:                &dep,
				Alias:                     args.Alias,
				DependencyVariantSelector: args.DepVariantSelector,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	addRequirement = task.WithContext(
		addRequirement,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			addRequirement,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while linking root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("add", "add a root as a dependency of the current root").
		WithArg(
			clib.
				NewArg("path", "path or root URL to add as a dependency (for example ../dep or root:../dep?arch=amd64&os=linux)").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(clib.NewArg("alias", "the alias to add the root under. if not specified, this defaults to the basename of the linked root").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
