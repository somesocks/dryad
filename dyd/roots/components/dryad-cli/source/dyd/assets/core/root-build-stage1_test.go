package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootBuildStage1DependencyName_ConcreteSingleTargetNoSuffix(t *testing.T) {
	assert := assert.New(t)

	err, descriptor := variantDescriptorParseFilesystem("os=linux")
	assert.Nil(err)

	err, name := rootBuild_stage1DependencyName(
		"foo",
		RootRequirementResolvedTarget{
			VariantDescriptor:  descriptor,
			ForceVariantSuffix: false,
		},
		1,
	)
	assert.Nil(err)
	assert.Equal("foo", name)
}

func TestRootBuildStage1DependencyName_AnySingleTargetAddsSuffix(t *testing.T) {
	assert := assert.New(t)

	err, descriptor := variantDescriptorParseFilesystem("os=linux")
	assert.Nil(err)

	err, name := rootBuild_stage1DependencyName(
		"foo",
		RootRequirementResolvedTarget{
			VariantDescriptor:  descriptor,
			ForceVariantSuffix: true,
		},
		1,
	)
	assert.Nil(err)
	assert.Equal("foo+os=linux", name)
}

func TestRootBuildStage1DependencyName_MultiTargetAddsSuffix(t *testing.T) {
	assert := assert.New(t)

	err, descriptor := variantDescriptorParseFilesystem("arch=amd64,os=linux")
	assert.Nil(err)

	err, name := rootBuild_stage1DependencyName(
		"foo",
		RootRequirementResolvedTarget{
			VariantDescriptor:  descriptor,
			ForceVariantSuffix: false,
		},
		2,
	)
	assert.Nil(err)
	assert.Equal("foo+arch=amd64,os=linux", name)
}
