package core

import (
	"dryad/internal/filepath"
	"dryad/task"
)

type SafeRootVariantAssetsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type SafeRootVariantCommandsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type SafeRootVariantTraitsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type SafeRootVariantSecretsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type SafeRootVariantDocsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type SafeRootVariantRequirementsReference struct {
	BasePath string
	Variant  *SafeRootVariantReference
}

type rootVariantSelectedPaths struct {
	Assets       *SafeRootVariantAssetsReference
	Commands     *SafeRootVariantCommandsReference
	Traits       *SafeRootVariantTraitsReference
	Secrets      *SafeRootVariantSecretsReference
	Docs         *SafeRootVariantDocsReference
	Requirements *SafeRootVariantRequirementsReference
}

func rootVariantSelectedAssets(path string, variant *SafeRootVariantReference) *SafeRootVariantAssetsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantAssetsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func rootVariantSelectedCommands(path string, variant *SafeRootVariantReference) *SafeRootVariantCommandsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantCommandsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func rootVariantSelectedTraits(path string, variant *SafeRootVariantReference) *SafeRootVariantTraitsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantTraitsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func rootVariantSelectedSecrets(path string, variant *SafeRootVariantReference) *SafeRootVariantSecretsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantSecretsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func rootVariantSelectedDocs(path string, variant *SafeRootVariantReference) *SafeRootVariantDocsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantDocsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func rootVariantSelectedRequirements(path string, variant *SafeRootVariantReference) *SafeRootVariantRequirementsReference {
	if path == "" {
		return nil
	}

	return &SafeRootVariantRequirementsReference{
		BasePath: path,
		Variant:  variant,
	}
}

func (requirements *SafeRootVariantRequirementsReference) Walk(
	ctx *task.ExecutionContext,
	req RootRequirementsWalkRequest,
) error {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.Walk(ctx, req)
}

func (requirements *SafeRootVariantRequirementsReference) Requirement(
	path string,
) *UnsafeRootRequirementReference {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.Requirement(path)
}

func (requirements *SafeRootVariantRequirementsReference) Add(
	ctx *task.ExecutionContext,
	req RootRequirementsAddRequest,
) (error, *SafeRootRequirementReference) {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.Add(ctx, req)
}

func (requirements *SafeRootVariantRequirementsReference) AddEnv(
	ctx *task.ExecutionContext,
	req RootRequirementsAddEnvRequest,
) (error, *SafeRootRequirementReference) {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.AddEnv(ctx, req)
}

func (requirements *SafeRootVariantRequirementsReference) AddFile(
	ctx *task.ExecutionContext,
	req RootRequirementsAddFileRequest,
) (error, *SafeRootRequirementReference) {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.AddFile(ctx, req)
}

func (requirements *SafeRootVariantRequirementsReference) AddHTTP(
	ctx *task.ExecutionContext,
	req RootRequirementsAddHTTPRequest,
) (error, *SafeRootRequirementReference) {
	rootRequirements := SafeRootRequirementsReference{
		BasePath: requirements.BasePath,
		Root:     requirements.Variant.Root,
	}

	return rootRequirements.AddHTTP(ctx, req)
}

func rootVariantExactSelectorDescriptor(
	ctx *task.ExecutionContext,
	variant *SafeRootVariantReference,
) (error, VariantDescriptor) {
	err, dimensions := variant.Root.VariantDimensions(ctx)
	if err != nil {
		return err, nil
	}

	if len(dimensions) == 0 {
		return nil, VariantDescriptor{}
	}

	selector := VariantDescriptor{}
	for _, dimension := range dimensions {
		option, exists := variant.Descriptor[dimension.Name]
		if exists {
			selector[dimension.Name] = option
			continue
		}

		selector[dimension.Name] = VariantOptionNone
	}

	return nil, selector
}

func (variant *SafeRootVariantReference) EnsureRequirements(
	ctx *task.ExecutionContext,
) (error, *SafeRootVariantRequirementsReference) {
	if variant.Requirements != nil {
		return nil, variant.Requirements
	}

	err, selector := rootVariantExactSelectorDescriptor(ctx, variant)
	if err != nil {
		return err, nil
	}

	requirementsPath := filepath.Join(variant.Root.BasePath, "dyd", "requirements")
	if len(selector) > 0 {
		err, selectorRaw := variantDescriptorEncodeFilesystem(selector)
		if err != nil {
			return err, nil
		}
		requirementsPath += RootRequirementSelectorSeparator + selectorRaw
	}

	err, requirements := (&UnsafeRootRequirementsReference{
		BasePath: requirementsPath,
		Root:     variant.Root,
	}).Ensure(ctx)
	if err != nil {
		return err, nil
	}

	variant.Requirements = &SafeRootVariantRequirementsReference{
		BasePath: requirements.BasePath,
		Variant:  variant,
	}

	return nil, variant.Requirements
}
