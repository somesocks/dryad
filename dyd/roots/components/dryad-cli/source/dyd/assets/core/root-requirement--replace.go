package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"fmt"

	"dryad/internal/filepath"
	"dryad/internal/os"
	"net/url"
	"strings"
	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

type RootReplaceTargetSpec struct {
	Root               *SafeRootReference
	VariantSelector    VariantDescriptor
	HasVariantSelector bool
}

func rootRequirementTargetSpecMatchesReplaceTarget(
	targetSpec *RootRequirementTargetSpec,
	matchSpec RootReplaceTargetSpec,
) bool {
	if rootRequirementTargetKind(targetSpec.Kind) != RootRequirementTargetKindRoot || targetSpec.Root == nil {
		return false
	}

	if targetSpec.Root.BasePath != matchSpec.Root.BasePath {
		return false
	}

	if !matchSpec.HasVariantSelector {
		return true
	}

	for dimension, option := range matchSpec.VariantSelector {
		targetOption, ok := targetSpec.VariantSelector[dimension]
		if !ok || targetOption != option {
			return false
		}
	}

	return true
}

func rootRequirementTargetSpecApplyReplaceTarget(
	targetSpec *RootRequirementTargetSpec,
	replaceSpec RootReplaceTargetSpec,
) (error, *RootRequirementTargetSpec) {
	if rootRequirementTargetKind(targetSpec.Kind) != RootRequirementTargetKindRoot {
		return nil, targetSpec
	}

	if replaceSpec.Root == nil {
		return fmt.Errorf("missing replacement root"), nil
	}

	nextSelector := VariantDescriptor{}
	for dimension, option := range targetSpec.VariantSelector {
		nextSelector[dimension] = option
	}

	if replaceSpec.HasVariantSelector {
		for dimension, option := range replaceSpec.VariantSelector {
			nextSelector[dimension] = option
		}
	}

	return nil, &RootRequirementTargetSpec{
		Kind:            RootRequirementTargetKindRoot,
		Root:            replaceSpec.Root,
		VariantSelector: nextSelector,
	}
}

func (rootRequirement *SafeRootRequirementReference) Replace(ctx *task.ExecutionContext, targetSpec *RootRequirementTargetSpec) error {
	var err error
	var linkTarget string

	if targetSpec == nil || targetSpec.Root == nil {
		return fmt.Errorf("missing replacement target")
	}
	if rootRequirementTargetKind(targetSpec.Kind) != RootRequirementTargetKindRoot {
		return fmt.Errorf("replacement target must be a root")
	}

	err, variantSelector := variantDescriptorEncodeURL(targetSpec.VariantSelector)
	if err != nil {
		return err
	}

	linkTarget, err = filepath.Rel(
		filepath.Dir(rootRequirement.BasePath),
		targetSpec.Root.BasePath)
	if err != nil {
		return err
	}

	var linkUrl url.URL = url.URL{
		Scheme: "root",
		Opaque: linkTarget,
	}
	linkUrl.RawQuery = strings.TrimPrefix(variantSelector, "?")

	err, _ = fs2.Remove(ctx, rootRequirement.BasePath)
	if err != nil {
		return err
	}

	err = os.WriteFile(rootRequirement.BasePath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}
