package core

import (
	"dryad/internal/filepath"
	"dryad/task"
)

func (root *SafeRootReference) Variants() *UnsafeRootVariantsReference {
	var rootVariantsRef = UnsafeRootVariantsReference{
		BasePath: filepath.Join(root.BasePath, "dyd", "variants"),
		Root:     root,
	}
	return &rootVariantsRef
}

func (root *SafeRootReference) VariantDimensions(ctx *task.ExecutionContext) (error, []VariantDimension) {
	err, variants := root.Variants().Resolve(ctx)
	if err != nil {
		return err, nil
	}
	if variants == nil {
		return nil, []VariantDimension{}
	}
	return variants.Dimensions(ctx)
}

func (root *SafeRootReference) VariantExclusions(ctx *task.ExecutionContext) (error, []VariantExclusion) {
	err, variants := root.Variants().Resolve(ctx)
	if err != nil {
		return err, nil
	}
	if variants == nil {
		return nil, []VariantExclusion{}
	}
	return variants.Exclusions(ctx)
}

func (root *SafeRootReference) VariantInclusions(ctx *task.ExecutionContext) (error, []VariantInclusion) {
	err, variants := root.Variants().Resolve(ctx)
	if err != nil {
		return err, nil
	}
	if variants == nil {
		return nil, []VariantInclusion{}
	}
	return variants.Inclusions(ctx)
}

func (root *SafeRootReference) variantMetadata(ctx *task.ExecutionContext) (error, []VariantDimension, []VariantInclusion, []VariantExclusion) {
	err, variants := root.Variants().Resolve(ctx)
	if err != nil {
		return err, nil, nil, nil
	}
	if variants == nil {
		return nil, []VariantDimension{}, []VariantInclusion{}, []VariantExclusion{}
	}

	err, dimensions := variants.Dimensions(ctx)
	if err != nil {
		return err, nil, nil, nil
	}
	err, inclusions := variants.Inclusions(ctx)
	if err != nil {
		return err, nil, nil, nil
	}
	err, exclusions := variants.Exclusions(ctx)
	if err != nil {
		return err, nil, nil, nil
	}

	return nil, dimensions, inclusions, exclusions
}
