package core

import (
	"dryad/task"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createRootForVariantRequirementTest(t *testing.T, gardenPath string, relRootPath string) string {
	t.Helper()

	rootPath := filepath.Join(gardenPath, "dyd", "roots", relRootPath)
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "type"), "root")
	return rootPath
}

func createRequirementForVariantRequirementTest(
	t *testing.T,
	sourceRootPath string,
	alias string,
	targetRootPath string,
	urlSuffix string,
) {
	t.Helper()

	requirementsPath := filepath.Join(sourceRootPath, "dyd", "requirements")
	relTargetPath, err := filepath.Rel(requirementsPath, targetRootPath)
	assert.Nil(t, err)

	writeFileForTest(
		t,
		filepath.Join(requirementsPath, alias),
		"root:"+relTargetPath+urlSuffix,
	)
}

func resolveRequirementForVariantRequirementTest(
	t *testing.T,
	gardenPath string,
	sourceRootPath string,
	alias string,
) *SafeRootRequirementReference {
	t.Helper()

	garden := &SafeGardenReference{BasePath: gardenPath}
	roots := &SafeRootsReference{
		BasePath: filepath.Join(gardenPath, "dyd", "roots"),
		Garden:   garden,
	}
	sourceRoot := SafeRootReference{
		BasePath: sourceRootPath,
		Roots:    roots,
	}

	err, requirementsRef := sourceRoot.Requirements().Resolve(task.SERIAL_CONTEXT)
	assert.Nil(t, err)

	err, requirement := requirementsRef.Requirement(alias).Resolve(task.SERIAL_CONTEXT)
	assert.Nil(t, err)
	assert.NotNil(t, requirement)

	return requirement
}

func TestRootRequirementTargetSpec_ParsesVariantSelector(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")
	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?os=linux&arch=amd64",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")

	err, targetSpec := requirement.TargetSpec(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(targetRootPath, targetSpec.Root.BasePath)

	err, selector := variantDescriptorEncodeFilesystem(targetSpec.VariantSelector)
	assert.Nil(err)
	assert.Equal("arch=amd64+os=linux", selector)
}

func TestRootRequirementTargetSpec_FragmentVariantSelectorFails(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")
	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?os=linux#arch=amd64",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")

	err, _ := requirement.TargetSpec(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "variant descriptor fragments are not supported")
}

func TestRootRequirementResolveTargets_InheritAndConcrete(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=amd64&os=inherit",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, parentVariant := variantDescriptorParseFilesystem("os=darwin")
	assert.Nil(err)

	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: parentVariant,
	})
	assert.Nil(err)
	assert.Len(targets, 1)
	assert.False(targets[0].ForceVariantSuffix)

	err, variant := variantDescriptorEncodeFilesystem(targets[0].VariantDescriptor)
	assert.Nil(err)
	assert.Equal("arch=amd64+os=darwin", variant)
}

func TestRootRequirementResolveTargets_InheritNoneFromParent(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "none"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?os=inherit",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 1)

	err, variant := variantDescriptorEncodeFilesystem(targets[0].VariantDescriptor)
	assert.Nil(err)
	assert.Equal("", variant)
}

func TestRootRequirementResolveTargets_AnyExpandsCartesianProduct(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 4)
	for _, target := range targets {
		assert.True(target.ForceVariantSuffix)
	}

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
		"arch=arm64+os=darwin",
		"arch=arm64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_AnySingleTargetStillForcesSuffix(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 1)
	assert.True(targets[0].ForceVariantSuffix)

	err, variant := variantDescriptorEncodeFilesystem(targets[0].VariantDescriptor)
	assert.Nil(err)
	assert.Equal("os=linux", variant)
}

func TestRootRequirementResolveTargets_OptionListExpandsSetAndForcesSuffix(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=amd64,arm64&os=linux",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 2)
	for _, target := range targets {
		assert.True(target.ForceVariantSuffix)
	}

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=linux",
		"arch=arm64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_UnderspecifiedFails(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=amd64",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, _ := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "under-specified requirement variant dimension")
}

func TestRootRequirementResolveTargets_HostResolvesCurrentRuntime(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", runtime.GOOS), "true")
	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?os=host",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 1)

	err, variant := variantDescriptorEncodeFilesystem(targets[0].VariantDescriptor)
	assert.Nil(err)
	assert.Equal("os="+runtime.GOOS, variant)
}

func TestRootRequirementResolveTargets_HostUnsupportedDimensionFails(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "version", "1.0"), "true")
	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?version=host",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, _ := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is only supported")
}

func TestRootRequirementResolveTargets_ExclusionsFilterResolvedVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=darwin"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 3)

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=linux",
		"arch=arm64+os=darwin",
		"arch=arm64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_ExclusionsAnySelectorFiltersResolvedVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_exclude", "arch=any+os=darwin"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 2)

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=linux",
		"arch=arm64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_ExclusionsPartialSelectorFiltersResolvedVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_exclude", "arch=arm64"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 2)

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_InclusionsFilterResolvedVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=any"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 2)

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
	}, variants)
}

func TestRootRequirementResolveTargets_InclusionsPartialSelectorFiltersResolvedVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_include", "os=darwin"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 2)

	variants := make([]string, 0, len(targets))
	for _, target := range targets {
		err, variant := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		assert.Nil(err)
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=arm64+os=darwin",
	}, variants)
}

func TestRootRequirementResolveTargets_EmptyInclusionMapIncludesAllVariants(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=linux"), "false")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=any&os=any",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, targets := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.Nil(err)
	assert.Len(targets, 4)
}

func TestRootRequirementResolveTargets_FailsWhenSelectionIsExcluded(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	sourceRootPath := createRootForVariantRequirementTest(t, gardenPath, "source")
	targetRootPath := createRootForVariantRequirementTest(t, gardenPath, "dep")

	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(targetRootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=darwin"), "true")

	createRequirementForVariantRequirementTest(
		t,
		sourceRootPath,
		"dep",
		targetRootPath,
		"?arch=amd64&os=darwin",
	)

	requirement := resolveRequirementForVariantRequirementTest(t, gardenPath, sourceRootPath, "dep")
	err, _ := requirement.ResolveTargets(task.SERIAL_CONTEXT, RootRequirementResolveTargetsRequest{
		ParentVariant: VariantDescriptor{},
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "resolved requirement variants are filtered by variants/_include and variants/_exclude")
}
