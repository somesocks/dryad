package core

import (
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func variantDescriptorFromFilesystemForTest(t *testing.T, raw string) VariantDescriptor {
	t.Helper()

	err, descriptor := variantDescriptorParseFilesystem(raw)
	assert.Nil(t, err)
	return descriptor
}

func encodeVariantDescriptorsForTest(t *testing.T, variants []VariantDescriptor) []string {
	t.Helper()

	encoded := make([]string, 0, len(variants))
	for _, variant := range variants {
		err, raw := variantDescriptorEncodeFilesystem(variant)
		assert.Nil(t, err)
		encoded = append(encoded, raw)
	}
	sort.Strings(encoded)
	return encoded
}

func TestRootResolveBuildVariants_DefaultsToAllEnabledOptions(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
		"arch=arm64+os=darwin",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_UnderspecifiedSelectorDefaultsMissingDimensionsToAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "arch=amd64"),
	})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_OptionListSelectorExpandsSet(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "arch=amd64,arm64+os=linux"),
	})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=linux",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_IgnoreUnknownDimensionsWhenRequested(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector:                variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
		IgnoreUnknownDimensions: true,
	})
	assert.Nil(err)

	assert.Equal([]string{
		"os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_RejectUnknownDimensionsWhenNotIgnored(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "over-specified root build variant dimension")
}

func TestRootResolveBuildVariants_NoneOptionOmitsDimension(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64",
		"arch=amd64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_DisabledOptionFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "false")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "os=linux"),
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "disabled root build variant option")
}

func TestRootResolveBuildVariants_InheritIsRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "os=inherit"),
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported")
}

func TestRootResolveBuildVariants_HostResolvesRuntime(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", runtime.GOOS), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "other"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "os=host"),
	})
	assert.Nil(err)
	assert.Equal([]string{
		"os=" + runtime.GOOS,
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_HostUnsupportedDimensionRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "tool", "local"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "tool=host"),
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is only supported")
}

func TestRootResolveBuildVariants_AppliesExclusions(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=linux"), "false")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=arm64+os=darwin"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_AppliesAnyExclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=any+os=darwin"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=linux",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_AppliesOptionListExclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=darwin,linux"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=arm64+os=darwin",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_AppliesPartialExclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=arm64"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_RejectsInheritInExclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=inherit"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for excluded variant selectors")
}

func TestRootResolveBuildVariants_RejectsHostInExclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", runtime.GOOS), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=host"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for excluded variant selectors")
}

func TestRootResolveBuildVariants_AppliesInclusions(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=arm64+os=darwin"), "false")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_AppliesPartialInclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "os=darwin"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=arm64+os=darwin",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_EmptyInclusionMapIncludesAllVariants(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=linux"), "false")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
		"arch=amd64+os=linux",
		"arch=arm64+os=darwin",
		"arch=arm64+os=linux",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_AppliesInclusionsAndExclusions(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=any+os=darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=arm64+os=darwin"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, variants := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.Nil(err)

	assert.Equal([]string{
		"arch=amd64+os=darwin",
	}, encodeVariantDescriptorsForTest(t, variants))
}

func TestRootResolveBuildVariants_RejectsInheritInInclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=inherit"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for included variant selectors")
}

func TestRootResolveBuildVariants_RejectsHostInInclusionSelector(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", runtime.GOOS), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=host"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for included variant selectors")
}

func TestRootResolveBuildVariants_FailsWhenSelectionIsExcluded(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=linux"), "true")

	root := SafeRootReference{BasePath: rootPath}
	err, _ := root.ResolveBuildVariants(task.SERIAL_CONTEXT, RootResolveBuildVariantsRequest{
		Selector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "resolved root build variants are filtered by variants/_include and variants/_exclude")
}
