package core

import (
	"path/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootDevelopMaterializeVariantTraits_AppliesSelectionAndRemovesVariants(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "demo")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "darwin"), "true")

	err := rootDevelop_copyDir(
		task.SERIAL_CONTEXT,
		filepath.Join(rootPath, "dyd", "traits"),
		filepath.Join(workspacePath, "dyd", "traits"),
		rootDevelopCopyOptions{ApplyIgnore: false},
	)
	assert.Nil(err)

	err = rootDevelop_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=linux")
	assert.Nil(err)

	assert.Equal("linux", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "os")))
	assert.Equal("demo", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "name")))

	workspaceVariantsPath := filepath.Join(workspacePath, "dyd", "traits", "variants")
	workspaceVariantsExists, err := fileExists(workspaceVariantsPath)
	assert.Nil(err)
	assert.False(workspaceVariantsExists)

	rootVariantsPath := filepath.Join(rootPath, "dyd", "traits", "variants")
	rootVariantsExists, err := fileExists(rootVariantsPath)
	assert.Nil(err)
	assert.True(rootVariantsExists)
}

func TestRootDevelopMaterializeVariantTraits_NoVariantsRejectsDescriptor(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "demo")
	writeFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "name"), "demo")

	err := rootDevelop_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "root has no variant dimensions")
}
