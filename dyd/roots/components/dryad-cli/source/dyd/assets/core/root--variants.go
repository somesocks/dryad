package core

import (
	"dryad/task"
	"path/filepath"
)

func (root *SafeRootReference) Variants() *UnsafeRootVariantsReference {
	var rootVariantsRef = UnsafeRootVariantsReference{
		BasePath: filepath.Join(root.BasePath, "dyd", "traits", "variants"),
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
