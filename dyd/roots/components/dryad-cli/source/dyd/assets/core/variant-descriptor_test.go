package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariantDescriptorNormalizeFilesystem_SortsDimensions(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeFilesystem("os=linux+arch=amd64")
	assert.Nil(err)
	assert.Equal("arch=amd64+os=linux", normalized)
}

func TestVariantDescriptorNormalizeFilesystem_DuplicateDimensionFails(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeFilesystem("os=linux+os=darwin")
	assert.NotNil(err)
	assert.Contains(err.Error(), "duplicate variant dimension")
}

func TestVariantDescriptorNormalizeFilesystem_NormalizesOptionLists(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeFilesystem("os=linux,darwin,linux+arch=arm64,amd64")
	assert.Nil(err)
	assert.Equal("arch=amd64,arm64+os=darwin,linux", normalized)
}

func TestVariantDescriptorNormalizeFilesystem_MalformedOptionListFails(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeFilesystem("os=linux,+arch=amd64")
	assert.NotNil(err)
	assert.Contains(err.Error(), "malformed variant descriptor")
}

func TestVariantDescriptorNormalizeURL_SortsDimensions(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeURL("?os=linux&arch=amd64")
	assert.Nil(err)
	assert.Equal("?arch=amd64&os=linux", normalized)
}

func TestVariantDescriptorNormalizeURL_EmptyDescriptor(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeURL("")
	assert.Nil(err)
	assert.Equal("", normalized)
}

func TestVariantDescriptorNormalizeURL_InvalidCharactersFail(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeURL("?arch=amd64&os=lin/ux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid variant option")
}

func TestVariantDescriptorNormalizeURL_NormalizesOptionLists(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeURL("?os=linux,darwin,linux&arch=arm64,amd64")
	assert.Nil(err)
	assert.Equal("?arch=amd64,arm64&os=darwin,linux", normalized)
}

func TestVariantDescriptorNormalizeURL_FragmentSeparatorFails(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeURL("?os=linux#arch=amd64")
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid variant option")
}
