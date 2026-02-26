package core

import (
	"path/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootVariantsInclusionsLoad_Basic(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=amd64+os=linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "arch=arm64+os=darwin"), "false")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, inclusions := root.VariantInclusions(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(
		[]VariantInclusion{
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
		inclusions,
	)
}

func TestRootVariantsInclusionsLoad_RejectsNonCanonicalDescriptor(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "_include", "os=linux+arch=amd64"), "true")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, _ := root.VariantInclusions(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "included variant descriptor must be canonical")
}
