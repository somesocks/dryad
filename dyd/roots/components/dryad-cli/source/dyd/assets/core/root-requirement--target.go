package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"net/url"
	"strings"
)

type RootRequirementTargetSpec struct {
	Root            *SafeRootReference
	VariantSelector VariantDescriptor
}

func rootRequirementVariantSelectorFromURL(linkURL *url.URL) (error, VariantDescriptor) {
	hasQuery := linkURL.RawQuery != ""
	hasFragment := linkURL.Fragment != ""

	if hasFragment {
		return fmt.Errorf("variant descriptor fragments are not supported; use query parameters with '&'"), nil
	}

	if !hasQuery && !hasFragment {
		return nil, VariantDescriptor{}
	}

	selectorRaw := "?"
	if hasQuery {
		selectorRaw += linkURL.RawQuery
	}

	return variantDescriptorParseURL(selectorRaw)
}

func (rootRequirement *SafeRootRequirementReference) TargetSpec(ctx *task.ExecutionContext) (error, *RootRequirementTargetSpec) {
	linkInfo, err := os.Lstat(rootRequirement.BasePath)
	if err != nil {
		return err, nil
	}

	isSymlink := linkInfo.Mode()&os.ModeSymlink == os.ModeSymlink

	// compatibility with legacy gardens - if a requirement is a symlink
	// resolve it directly
	if isSymlink {
		linkPath, err := os.Readlink(rootRequirement.BasePath)
		if err != nil {
			return err, nil
		}

		// convert relative links to an absolute path
		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(
				filepath.Dir(rootRequirement.BasePath),
				linkPath,
			)
		}

		err, safeRef := rootRequirement.Requirements.Root.Roots.Root(linkPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		return nil, &RootRequirementTargetSpec{
			Root:            &safeRef,
			VariantSelector: VariantDescriptor{},
		}
	}

	linkBytes, err := os.ReadFile(rootRequirement.BasePath)
	if err != nil {
		return err, nil
	}

	linkRaw := string(linkBytes)
	linkString := strings.TrimSpace(linkRaw)
	warnRequirementFileWhitespace(
		sentinelLogPath(
			rootRequirement.BasePath,
			rootRequirement.Requirements.Root.Roots.Garden.BasePath,
		),
		linkRaw,
		linkString,
	)

	linkURL, err := url.Parse(linkString)
	if err != nil {
		return err, nil
	}

	linkScheme := linkURL.Scheme
	if linkScheme != "root" {
		return fmt.Errorf("unsupported scheme for root requirement: %s", linkScheme), nil
	}

	// check for an opaque path, otherwise use regular
	linkPath := linkURL.Opaque
	if linkPath == "" {
		linkPath = linkURL.Path
	}

	if filepath.IsAbs(linkPath) {
		return fmt.Errorf("root requirement path must be relative: %s", linkPath), nil
	}

	linkPath = filepath.Join(
		filepath.Dir(rootRequirement.BasePath),
		linkPath,
	)

	err, safeRef := rootRequirement.Requirements.Root.Roots.Root(linkPath).Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, selector := rootRequirementVariantSelectorFromURL(linkURL)
	if err != nil {
		return err, nil
	}

	return nil, &RootRequirementTargetSpec{
		Root:            &safeRef,
		VariantSelector: selector,
	}
}

func (rootRequirement *SafeRootRequirementReference) Target(ctx *task.ExecutionContext) (error, *SafeRootReference) {
	err, targetSpec := rootRequirement.TargetSpec(ctx)
	if err != nil {
		return err, nil
	}
	return nil, targetSpec.Root
}
