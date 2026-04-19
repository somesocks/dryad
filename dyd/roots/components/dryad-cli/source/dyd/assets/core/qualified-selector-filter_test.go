package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQualifiedSelector_PathOnly(t *testing.T) {
	err, selector := parseQualifiedSelector("dyd/roots/**")

	assert.NoError(t, err)
	assert.Equal(t, "dyd/roots/**", selector.PathGlob)
	assert.False(t, selector.HasSelector)
	assert.Equal(t, VariantDescriptor{}, selector.Descriptor)
}

func TestParseQualifiedSelector_PathAndSelector(t *testing.T) {
	err, selector := parseQualifiedSelector("dyd/roots/**~os=linux")

	assert.NoError(t, err)
	assert.Equal(t, "dyd/roots/**", selector.PathGlob)
	assert.True(t, selector.HasSelector)
	assert.Equal(t, VariantDescriptor{"os": "linux"}, selector.Descriptor)
}

func TestParseQualifiedSelector_NakedSelectorDefaultsPathGlob(t *testing.T) {
	err, selector := parseQualifiedSelector("~os=linux")

	assert.NoError(t, err)
	assert.Equal(t, "**", selector.PathGlob)
	assert.True(t, selector.HasSelector)
	assert.Equal(t, VariantDescriptor{"os": "linux"}, selector.Descriptor)
}

func TestParseQualifiedSelector_RejectsEmptyNakedSelector(t *testing.T) {
	err, _ := parseQualifiedSelector("~")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector descriptor is empty")
}
