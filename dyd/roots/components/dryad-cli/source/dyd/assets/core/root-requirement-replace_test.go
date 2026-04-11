package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootRequirementTargetSpecMatchesReplaceTarget_UnqualifiedMatchesSameRoot(t *testing.T) {
	assert := assert.New(t)

	sourceRoot := &SafeRootReference{BasePath: "/tmp/source"}
	otherRoot := &SafeRootReference{BasePath: "/tmp/other"}
	targetSpec := &RootRequirementTargetSpec{
		Root:            sourceRoot,
		VariantSelector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	}

	assert.True(rootRequirementTargetSpecMatchesReplaceTarget(
		targetSpec,
		RootReplaceTargetSpec{
			Root: sourceRoot,
		},
	))
	assert.False(rootRequirementTargetSpecMatchesReplaceTarget(
		targetSpec,
		RootReplaceTargetSpec{
			Root: otherRoot,
		},
	))
}

func TestRootRequirementTargetSpecMatchesReplaceTarget_QualifiedRequiresExplicitMatch(t *testing.T) {
	assert := assert.New(t)

	sourceRoot := &SafeRootReference{BasePath: "/tmp/source"}
	targetSpec := &RootRequirementTargetSpec{
		Root:            sourceRoot,
		VariantSelector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	}
	targetWithoutOS := &RootRequirementTargetSpec{
		Root:            sourceRoot,
		VariantSelector: variantDescriptorFromFilesystemForTest(t, "arch=amd64"),
	}

	assert.True(rootRequirementTargetSpecMatchesReplaceTarget(
		targetSpec,
		RootReplaceTargetSpec{
			Root:               sourceRoot,
			VariantSelector:    variantDescriptorFromFilesystemForTest(t, "os=linux"),
			HasVariantSelector: true,
		},
	))
	assert.False(rootRequirementTargetSpecMatchesReplaceTarget(
		targetWithoutOS,
		RootReplaceTargetSpec{
			Root:               sourceRoot,
			VariantSelector:    variantDescriptorFromFilesystemForTest(t, "os=linux"),
			HasVariantSelector: true,
		},
	))
}

func TestRootRequirementTargetSpecApplyReplaceTarget_PreservesUnspecifiedDimensions(t *testing.T) {
	assert := assert.New(t)

	oldRoot := &SafeRootReference{BasePath: "/tmp/old"}
	newRoot := &SafeRootReference{BasePath: "/tmp/new"}
	targetSpec := &RootRequirementTargetSpec{
		Root:            oldRoot,
		VariantSelector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	}

	err, replacedTarget := rootRequirementTargetSpecApplyReplaceTarget(
		targetSpec,
		RootReplaceTargetSpec{
			Root:               newRoot,
			VariantSelector:    variantDescriptorFromFilesystemForTest(t, "os=darwin"),
			HasVariantSelector: true,
		},
	)
	assert.Nil(err)
	assert.Equal(newRoot.BasePath, replacedTarget.Root.BasePath)

	err, selector := variantDescriptorEncodeFilesystem(replacedTarget.VariantSelector)
	assert.Nil(err)
	assert.Equal("arch=amd64+os=darwin", selector)
}

func TestRootRequirementTargetSpecApplyReplaceTarget_UnqualifiedPreservesSelector(t *testing.T) {
	assert := assert.New(t)

	oldRoot := &SafeRootReference{BasePath: "/tmp/old"}
	newRoot := &SafeRootReference{BasePath: "/tmp/new"}
	targetSpec := &RootRequirementTargetSpec{
		Root:            oldRoot,
		VariantSelector: variantDescriptorFromFilesystemForTest(t, "arch=amd64+os=linux"),
	}

	err, replacedTarget := rootRequirementTargetSpecApplyReplaceTarget(
		targetSpec,
		RootReplaceTargetSpec{
			Root: newRoot,
		},
	)
	assert.Nil(err)
	assert.Equal(newRoot.BasePath, replacedTarget.Root.BasePath)

	err, selector := variantDescriptorEncodeFilesystem(replacedTarget.VariantSelector)
	assert.Nil(err)
	assert.Equal("arch=amd64+os=linux", selector)
}
