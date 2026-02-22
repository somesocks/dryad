package core

import (
	"fmt"
	"regexp"
	"strings"
)

const RootRequirementSelectorSeparator = "~"

var ROOT_REQUIREMENT_ALIAS_RE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func rootRequirementParseName(raw string) (error, string, VariantDescriptor) {
	alias, conditionRaw, hasCondition := strings.Cut(raw, RootRequirementSelectorSeparator)
	if alias == "" {
		return fmt.Errorf("malformed requirement name: %s", raw), "", nil
	}
	if !ROOT_REQUIREMENT_ALIAS_RE.MatchString(alias) {
		return fmt.Errorf("malformed requirement name: %s", raw), "", nil
	}

	if !hasCondition {
		return nil, alias, VariantDescriptor{}
	}
	if conditionRaw == "" {
		return fmt.Errorf("malformed requirement condition descriptor: %s", raw), "", nil
	}

	err, condition := variantDescriptorParseFilesystem(conditionRaw)
	if err != nil {
		return fmt.Errorf("malformed requirement condition descriptor: %s", raw), "", nil
	}

	return nil, alias, condition
}

func rootRequirementEncodeName(alias string, condition VariantDescriptor) (error, string) {
	if !ROOT_REQUIREMENT_ALIAS_RE.MatchString(alias) {
		return fmt.Errorf("malformed requirement name: %s", alias), ""
	}

	if len(condition) == 0 {
		return nil, alias
	}

	err, conditionRaw := variantDescriptorEncodeFilesystem(condition)
	if err != nil {
		return err, ""
	}

	return nil, alias + RootRequirementSelectorSeparator + conditionRaw
}

func RootRequirementNormalizeName(raw string) (error, string) {
	err, alias, condition := rootRequirementParseName(raw)
	if err != nil {
		return err, ""
	}

	return rootRequirementEncodeName(alias, condition)
}

func rootRequirementConditionMatches(
	parentVariant VariantDescriptor,
	condition VariantDescriptor,
) (error, bool) {
	for dimension, option := range condition {
		switch option {
		case VariantOptionAny, VariantOptionInherit:
			continue

		case VariantOptionHost:
			err, hostOption := rootRequirementHostOption(dimension)
			if err != nil {
				return err, false
			}

			parentOption, exists := parentVariant[dimension]
			if !exists || parentOption != hostOption {
				return nil, false
			}

		case VariantOptionNone:
			_, exists := parentVariant[dimension]
			if exists {
				return nil, false
			}

		default:
			parentOption, exists := parentVariant[dimension]
			if !exists || parentOption != option {
				return nil, false
			}
		}
	}

	return nil, true
}
