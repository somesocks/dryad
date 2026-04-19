package core

import (
	"fmt"
	"strings"

	"dryad/internal/filepath"
	"dryad/task"

	"github.com/bmatcuk/doublestar/v4"
)

type qualifiedSelector struct {
	PathGlob    string
	Descriptor  VariantDescriptor
	HasSelector bool
}

type SelectorFilterRequest struct {
	Include []string
	Exclude []string
}

func parseQualifiedSelector(raw string) (error, qualifiedSelector) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty selector"), qualifiedSelector{}
	}

	pathGlob := raw
	selectorRaw := ""
	hasSelector := false

	if strings.HasPrefix(raw, RootRequirementSelectorSeparator) {
		pathGlob = "**"
		selectorRaw = raw[len(RootRequirementSelectorSeparator):]
		hasSelector = true
	} else if splitIndex := strings.LastIndex(raw, RootRequirementSelectorSeparator); splitIndex >= 0 {
		pathGlob = raw[:splitIndex]
		selectorRaw = raw[splitIndex+len(RootRequirementSelectorSeparator):]
		hasSelector = true
	}

	if pathGlob == "" {
		return fmt.Errorf("selector path glob is empty: %s", raw), qualifiedSelector{}
	}

	if !doublestar.ValidatePattern(pathGlob) {
		return fmt.Errorf("invalid selector path glob: %s", pathGlob), qualifiedSelector{}
	}

	selector := VariantDescriptor{}
	if hasSelector {
		if selectorRaw == "" {
			return fmt.Errorf("selector descriptor is empty: %s", raw), qualifiedSelector{}
		}

		err, parsed := variantDescriptorParseFilesystem(selectorRaw)
		if err != nil {
			return fmt.Errorf("malformed selector descriptor: %s", raw), qualifiedSelector{}
		}
		selector = parsed
	}

	return nil, qualifiedSelector{
		PathGlob:    pathGlob,
		Descriptor:  selector,
		HasSelector: hasSelector,
	}
}

func parseQualifiedSelectors(rawSelectors []string) (error, []qualifiedSelector) {
	selectors := make([]qualifiedSelector, 0, len(rawSelectors))
	for _, rawSelector := range rawSelectors {
		err, selector := parseQualifiedSelector(rawSelector)
		if err != nil {
			return err, nil
		}
		selectors = append(selectors, selector)
	}

	return nil, selectors
}

func qualifiedSelectorPathMatches(selector qualifiedSelector, path string) (error, bool) {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "./")

	matched, err := doublestar.Match(selector.PathGlob, path)
	if err != nil {
		return err, false
	}

	return nil, matched
}

func selectorOptionMatchesTraitValue(requestedOption string, value string, exists bool) bool {
	switch requestedOption {
	case VariantOptionAny:
		return exists
	case VariantOptionNone:
		return !exists
	default:
		return exists && strings.TrimSpace(value) == requestedOption
	}
}

func selectorOptionsMatchTraitValue(requestedOptionRaw string, value string, exists bool) (error, bool) {
	err, requestedOptions := variantDescriptorOptionValues(requestedOptionRaw)
	if err != nil {
		return err, false
	}

	for _, requestedOption := range requestedOptions {
		if selectorOptionMatchesTraitValue(requestedOption, value, exists) {
			return nil, true
		}
	}

	return nil, false
}

func selectorOptionMatchesVariantValue(requestedOption string, value string, exists bool) bool {
	switch requestedOption {
	case VariantOptionAny:
		return exists
	case VariantOptionNone:
		return !exists
	default:
		return exists && value == requestedOption
	}
}

func selectorOptionsMatchVariantValue(requestedOptionRaw string, value string, exists bool) (error, bool) {
	err, requestedOptions := variantDescriptorOptionValues(requestedOptionRaw)
	if err != nil {
		return err, false
	}

	for _, requestedOption := range requestedOptions {
		if selectorOptionMatchesVariantValue(requestedOption, value, exists) {
			return nil, true
		}
	}

	return nil, false
}

func rootVariantGardenPath(variant *SafeRootVariantReference) (string, error) {
	return filepath.Rel(variant.Root.Roots.Garden.BasePath, variant.Root.BasePath)
}

func sproutGardenPath(sprout *SafeSproutReference) (string, error) {
	return filepath.Rel(sprout.Sprouts.Garden.BasePath, sprout.BasePath)
}

func qualifiedSelectorsMatchRootVariant(
	ctx *task.ExecutionContext,
	selectors []qualifiedSelector,
	variant *SafeRootVariantReference,
) (error, bool) {
	for _, selector := range selectors {
		err, matches := qualifiedSelectorMatchesRootVariant(ctx, selector, variant)
		if err != nil {
			return err, false
		}
		if matches {
			return nil, true
		}
	}

	return nil, false
}

func qualifiedSelectorsMatchSprout(
	ctx *task.ExecutionContext,
	selectors []qualifiedSelector,
	sprout *SafeSproutReference,
) (error, bool) {
	for _, selector := range selectors {
		err, matches := qualifiedSelectorMatchesSprout(ctx, selector, sprout)
		if err != nil {
			return err, false
		}
		if matches {
			return nil, true
		}
	}

	return nil, false
}
