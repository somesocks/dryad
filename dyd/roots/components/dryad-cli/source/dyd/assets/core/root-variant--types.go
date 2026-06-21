package core

import (
	"dryad/task"
	"fmt"
	"strings"
)

type UnsafeRootVariantReference struct {
	Root       *SafeRootReference
	Descriptor VariantDescriptor
}

type SafeRootVariantReference struct {
	Root         *SafeRootReference
	Descriptor   VariantDescriptor
	Dimensions   []VariantDimension
	Assets       *SafeRootVariantAssetsReference
	Commands     *SafeRootVariantCommandsReference
	Traits       *SafeRootVariantTraitsReference
	Secrets      *SafeRootVariantSecretsReference
	Docs         *SafeRootVariantDocsReference
	Requirements *SafeRootVariantRequirementsReference
}

func (root *SafeRootReference) Variant(
	descriptor VariantDescriptor,
) *UnsafeRootVariantReference {
	return &UnsafeRootVariantReference{
		Root:       root,
		Descriptor: descriptor,
	}
}

func (root *SafeRootReference) VariantFromFilesystem(
	raw string,
) (error, *UnsafeRootVariantReference) {
	err, context := RootVariantContextFromFilesystem(raw)
	if err != nil {
		return err, nil
	}

	return nil, root.Variant(context.Descriptor)
}

func (root *SafeRootReference) VariantFromURL(
	raw string,
) (error, *UnsafeRootVariantReference) {
	err, context := RootVariantContextFromURL(raw)
	if err != nil {
		return err, nil
	}

	return nil, root.Variant(context.Descriptor)
}

func (root *SafeRootReference) ResolveBuildVariantReferences(
	ctx *task.ExecutionContext,
	req RootResolveBuildVariantsRequest,
) (error, []*SafeRootVariantReference) {
	err, descriptors := root.ResolveBuildVariants(ctx, req)
	if err != nil {
		return err, nil
	}

	variants := make([]*SafeRootVariantReference, 0, len(descriptors))
	for _, descriptor := range descriptors {
		err, variant := root.Variant(descriptor).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		variants = append(variants, variant)
	}

	return nil, variants
}

func (root *SafeRootReference) ResolveBuildVariantReference(
	ctx *task.ExecutionContext,
	req RootResolveBuildVariantsRequest,
) (error, *SafeRootVariantReference) {
	err, variants := root.ResolveBuildVariantReferences(ctx, req)
	if err != nil {
		return err, nil
	}

	if len(variants) == 0 {
		return fmt.Errorf("no root variants resolved"), nil
	}

	if len(variants) > 1 {
		rendered := make([]string, 0, len(variants))
		for _, variant := range variants {
			err, descriptor := variant.Filesystem()
			if err != nil {
				return err, nil
			}
			rendered = append(rendered, descriptor)
		}

		return fmt.Errorf(
			"under-specified root variant selector: resolved %d variants (%s)",
			len(variants),
			strings.Join(rendered, ", "),
		), nil
	}

	return nil, variants[0]
}

func (variant *SafeRootVariantReference) Context() RootVariantContext {
	return RootVariantContext{Descriptor: variant.Descriptor}
}

func (variant *SafeRootVariantReference) Filesystem() (error, string) {
	return variant.Context().Filesystem()
}

func (variant *SafeRootVariantReference) URL() (error, string) {
	return variant.Context().URL()
}
