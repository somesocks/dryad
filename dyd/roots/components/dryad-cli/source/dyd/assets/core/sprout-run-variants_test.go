package core

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSproutRunVariantDescriptor(t *testing.T, raw string) VariantDescriptor {
	t.Helper()

	err, descriptor := variantDescriptorParseFilesystem(raw)
	assert.Nil(t, err)
	return descriptor
}

func testSproutRunVariantDescriptors(t *testing.T, variants []sproutRunStemVariant) []string {
	t.Helper()

	raws := make([]string, 0, len(variants))
	for _, variant := range variants {
		raws = append(raws, variant.DescriptorRaw)
	}
	return raws
}

func TestResolveSproutRunStemVariants_AnySelectorReturnsSet(t *testing.T) {
	assert := assert.New(t)

	available := []sproutRunStemVariant{
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=amd64+os=linux"), DescriptorRaw: "arch=amd64+os=linux"},
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=arm64+os=linux"), DescriptorRaw: "arch=arm64+os=linux"},
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=amd64+os=darwin"), DescriptorRaw: "arch=amd64+os=darwin"},
	}

	err, selected := resolveSproutRunStemVariants(
		available,
		testSproutRunVariantDescriptor(t, "arch=any+os=linux"),
	)
	assert.Nil(err)
	assert.Equal(
		[]string{
			"arch=amd64+os=linux",
			"arch=arm64+os=linux",
		},
		testSproutRunVariantDescriptors(t, selected),
	)
}

func TestResolveSproutRunStemVariants_OptionListSelectorReturnsSet(t *testing.T) {
	assert := assert.New(t)

	available := []sproutRunStemVariant{
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=amd64+os=linux"), DescriptorRaw: "arch=amd64+os=linux"},
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=arm64+os=linux"), DescriptorRaw: "arch=arm64+os=linux"},
		{Descriptor: testSproutRunVariantDescriptor(t, "arch=amd64+os=darwin"), DescriptorRaw: "arch=amd64+os=darwin"},
	}

	err, selected := resolveSproutRunStemVariants(
		available,
		testSproutRunVariantDescriptor(t, "arch=amd64,arm64+os=linux"),
	)
	assert.Nil(err)
	assert.Equal(
		[]string{
			"arch=amd64+os=linux",
			"arch=arm64+os=linux",
		},
		testSproutRunVariantDescriptors(t, selected),
	)
}

func TestResolveSproutRunStemVariants_HostSelectorResolvesRuntime(t *testing.T) {
	assert := assert.New(t)

	available := []sproutRunStemVariant{
		{
			Descriptor:    VariantDescriptor{"os": runtime.GOOS},
			DescriptorRaw: "os=" + runtime.GOOS,
		},
	}

	err, selected := resolveSproutRunStemVariants(
		available,
		testSproutRunVariantDescriptor(t, "os=host"),
	)
	assert.Nil(err)
	assert.Len(selected, 1)
	assert.Equal("os="+runtime.GOOS, selected[0].DescriptorRaw)
}

func TestResolveSproutRunStemVariants_InheritSelectorIsRejected(t *testing.T) {
	assert := assert.New(t)

	available := []sproutRunStemVariant{
		{
			Descriptor:    VariantDescriptor{"os": "linux"},
			DescriptorRaw: "os=linux",
		},
	}

	err, _ := resolveSproutRunStemVariants(
		available,
		testSproutRunVariantDescriptor(t, "os=inherit"),
	)
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported")
}
