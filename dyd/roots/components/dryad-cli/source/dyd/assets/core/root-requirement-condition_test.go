package core

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootRequirementParseName_NoCondition(t *testing.T) {
	assert := assert.New(t)

	err, alias, condition := rootRequirementParseName("foo")
	assert.Nil(err)
	assert.Equal("foo", alias)
	assert.Equal(VariantDescriptor{}, condition)
}

func TestRootRequirementParseName_WithCondition(t *testing.T) {
	assert := assert.New(t)

	err, alias, condition := rootRequirementParseName("foo~arch=any+os=linux")
	assert.Nil(err)
	assert.Equal("foo", alias)
	assert.Equal(VariantDescriptor{
		"os":   "linux",
		"arch": "any",
	}, condition)
}

func TestRootRequirementParseName_WithConditionOptionLists(t *testing.T) {
	assert := assert.New(t)

	err, alias, condition := rootRequirementParseName("foo~os=linux,darwin+arch=arm64,amd64")
	assert.Nil(err)
	assert.Equal("foo", alias)
	assert.Equal(VariantDescriptor{
		"arch": "amd64,arm64",
		"os":   "darwin,linux",
	}, condition)
}

func TestRootRequirementParseName_InvalidConditionFails(t *testing.T) {
	assert := assert.New(t)

	err, _, _ := rootRequirementParseName("foo~")
	assert.NotNil(err)
	assert.Contains(err.Error(), "malformed requirement condition descriptor")
}

func TestRootRequirementParseName_InvalidAliasFails(t *testing.T) {
	assert := assert.New(t)

	err, _, _ := rootRequirementParseName("foo+arch=amd64")
	assert.NotNil(err)
	assert.Contains(err.Error(), "malformed requirement name")

	err, _, _ = rootRequirementParseName("foo!~arch=amd64")
	assert.NotNil(err)
	assert.Contains(err.Error(), "malformed requirement name")
}

func TestRootRequirementNormalizeName_NoCondition(t *testing.T) {
	assert := assert.New(t)

	err, normalized := RootRequirementNormalizeName("foo")
	assert.Nil(err)
	assert.Equal("foo", normalized)
}

func TestRootRequirementNormalizeName_WithConditionCanonicalizesOrder(t *testing.T) {
	assert := assert.New(t)

	err, normalized := RootRequirementNormalizeName("foo~os=linux+arch=any")
	assert.Nil(err)
	assert.Equal("foo~arch=any+os=linux", normalized)
}

func TestRootRequirementEncodeName_InvalidAliasFails(t *testing.T) {
	assert := assert.New(t)

	err, _ := rootRequirementEncodeName("foo+bar", VariantDescriptor{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "malformed requirement name")
}

func TestRootRequirementConditionMatches_ConcreteAnyAndNone(t *testing.T) {
	assert := assert.New(t)

	err, matches := rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": "amd64",
			"os":   "linux",
		},
		VariantDescriptor{
			"arch": "any",
			"os":   "linux",
		},
	)
	assert.Nil(err)
	assert.True(matches)

	err, matches = rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": "amd64",
			"os":   "linux",
		},
		VariantDescriptor{
			"arch": "arm64",
		},
	)
	assert.Nil(err)
	assert.False(matches)

	err, matches = rootRequirementConditionMatches(
		VariantDescriptor{},
		VariantDescriptor{
			"arch": "none",
		},
	)
	assert.Nil(err)
	assert.True(matches)

	err, matches = rootRequirementConditionMatches(
		VariantDescriptor{},
		VariantDescriptor{
			"arch": "any",
		},
	)
	assert.Nil(err)
	assert.True(matches)
}

func TestRootRequirementConditionMatches_InheritIsWildcard(t *testing.T) {
	assert := assert.New(t)

	err, matches := rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": "amd64",
		},
		VariantDescriptor{
			"arch": "inherit",
		},
	)
	assert.Nil(err)
	assert.True(matches)
}

func TestRootRequirementConditionMatches_OptionListMatchesAny(t *testing.T) {
	assert := assert.New(t)

	err, matches := rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": "amd64",
			"os":   "linux",
		},
		VariantDescriptor{
			"arch": "arm64,amd64",
			"os":   "darwin,linux",
		},
	)
	assert.Nil(err)
	assert.True(matches)
}

func TestRootRequirementConditionMatches_Host(t *testing.T) {
	assert := assert.New(t)

	err, matches := rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": runtime.GOARCH,
		},
		VariantDescriptor{
			"arch": "host",
		},
	)
	assert.Nil(err)
	assert.True(matches)

	err, matches = rootRequirementConditionMatches(
		VariantDescriptor{
			"arch": "amd64",
		},
		VariantDescriptor{
			"os": "host",
		},
	)
	assert.Nil(err)
	assert.False(matches)

	err, _ = rootRequirementConditionMatches(
		VariantDescriptor{
			"tool": "go",
		},
		VariantDescriptor{
			"tool": "host",
		},
	)
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is only supported")
}
