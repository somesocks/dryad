package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type VariantDescriptor map[string]string

func variantDescriptorParse(raw string, separator string, supportsLeadingQuestion bool) (error, VariantDescriptor) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, VariantDescriptor{}
	}

	if supportsLeadingQuestion && strings.HasPrefix(raw, "?") {
		raw = strings.TrimPrefix(raw, "?")
		if raw == "" {
			return errors.New("malformed variant descriptor"), nil
		}
	}

	segments := strings.Split(raw, separator)
	descriptor := VariantDescriptor{}
	for _, segment := range segments {
		if segment == "" {
			return errors.New("malformed variant descriptor"), nil
		}

		parts := strings.SplitN(segment, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("malformed variant descriptor"), nil
		}

		dimension := parts[0]
		option := parts[1]
		if !variantNameValid(dimension) {
			return fmt.Errorf("invalid variant dimension in descriptor: %s", dimension), nil
		}
		if !variantNameValid(option) {
			return fmt.Errorf("invalid variant option in descriptor: %s", option), nil
		}

		_, exists := descriptor[dimension]
		if exists {
			return fmt.Errorf("duplicate variant dimension in descriptor: %s", dimension), nil
		}
		descriptor[dimension] = option
	}

	return nil, descriptor
}

func variantDescriptorEncode(descriptor VariantDescriptor, separator string, prefix string) (error, string) {
	keys := make([]string, 0, len(descriptor))
	for dimension, option := range descriptor {
		if !variantNameValid(dimension) {
			return fmt.Errorf("invalid variant dimension in descriptor: %s", dimension), ""
		}
		if !variantNameValid(option) {
			return fmt.Errorf("invalid variant option in descriptor: %s", option), ""
		}
		keys = append(keys, dimension)
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return nil, ""
	}

	segments := make([]string, 0, len(keys))
	for _, key := range keys {
		segments = append(segments, key+"="+descriptor[key])
	}

	return nil, prefix + strings.Join(segments, separator)
}

func variantDescriptorParseFilesystem(raw string) (error, VariantDescriptor) {
	return variantDescriptorParse(raw, ",", false)
}

func variantDescriptorEncodeFilesystem(descriptor VariantDescriptor) (error, string) {
	return variantDescriptorEncode(descriptor, ",", "")
}

func variantDescriptorNormalizeFilesystem(raw string) (error, string) {
	err, descriptor := variantDescriptorParseFilesystem(raw)
	if err != nil {
		return err, ""
	}
	return variantDescriptorEncodeFilesystem(descriptor)
}

func variantDescriptorParseURL(raw string) (error, VariantDescriptor) {
	return variantDescriptorParse(raw, "#", true)
}

func variantDescriptorEncodeURL(descriptor VariantDescriptor) (error, string) {
	return variantDescriptorEncode(descriptor, "#", "?")
}

func variantDescriptorNormalizeURL(raw string) (error, string) {
	err, descriptor := variantDescriptorParseURL(raw)
	if err != nil {
		return err, ""
	}
	return variantDescriptorEncodeURL(descriptor)
}
