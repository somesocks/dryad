package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariantDescriptorNormalizeFilesystem_SortsDimensions(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeFilesystem("os=linux,arch=amd64")
	assert.Nil(err)
	assert.Equal("arch=amd64,os=linux", normalized)
}

func TestVariantDescriptorNormalizeFilesystem_DuplicateDimensionFails(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeFilesystem("os=linux,os=darwin")
	assert.NotNil(err)
	assert.Contains(err.Error(), "duplicate variant dimension")
}

func TestVariantDescriptorNormalizeURL_SortsDimensions(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeURL("?os=linux#arch=amd64")
	assert.Nil(err)
	assert.Equal("?arch=amd64#os=linux", normalized)
}

func TestVariantDescriptorNormalizeURL_EmptyDescriptor(t *testing.T) {
	assert := assert.New(t)

	err, normalized := variantDescriptorNormalizeURL("")
	assert.Nil(err)
	assert.Equal("", normalized)
}

func TestVariantDescriptorNormalizeURL_InvalidCharactersFail(t *testing.T) {
	assert := assert.New(t)

	err, _ := variantDescriptorNormalizeURL("?arch=amd64&os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid variant option")
}
