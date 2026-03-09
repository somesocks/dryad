package core

import (
	"dryad/internal/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootVariantsExclusionsLoad_Basic(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=amd64+os=linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "arch=arm64+os=darwin"), "false")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, exclusions := root.VariantExclusions(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(
		[]VariantExclusion{
			{
				Descriptor: VariantDescriptor{
					"arch": "amd64",
					"os":   "linux",
				},
				Enabled: true,
			},
			{
				Descriptor: VariantDescriptor{
					"arch": "arm64",
					"os":   "darwin",
				},
				Enabled: false,
			},
		},
		exclusions,
	)
}

func TestRootVariantsExclusionsLoad_RejectsNonCanonicalDescriptor(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_exclude", "os=linux+arch=amd64"), "true")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, _ := root.VariantExclusions(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "excluded variant descriptor must be canonical")
}
