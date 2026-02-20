package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func readTrimmedFileForTest(t *testing.T, path string) string {
	t.Helper()
	bytes, err := os.ReadFile(path)
	assert.Nil(t, err)
	return strings.TrimSpace(string(bytes))
}

func TestRootBuildMaterializeVariantTraits_AppliesSelectionAndRemovesVariants(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "demo")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "arch", "amd64"), "true")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "arch=amd64,os=linux")
	assert.Nil(err)

	assert.Equal("demo", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "name")))
	assert.Equal("amd64", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "arch")))
	assert.Equal("linux", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "os")))

	variantsPath := filepath.Join(workspacePath, "dyd", "traits", "variants")
	variantsExists, err := fileExists(variantsPath)
	assert.Nil(err)
	assert.False(variantsExists)
}

func TestRootBuildMaterializeVariantTraits_UnderspecifiedFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "arch", "amd64"), "true")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "under-specified root variant descriptor dimension")
}

func TestRootBuildMaterializeVariantTraits_NoVariantsRejectsDescriptor(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "demo")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "root has no variant dimensions")
}

func TestRootBuildMaterializeVariantTraits_NoVariantsKeepsSymlinkBehavior(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "demo")
	writeFileForTest(t, filepath.Join(workspacePath, "dyd", ".keep"), "")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "")
	assert.Nil(err)
	if err != nil {
		return
	}

	workspaceTraitsPath := filepath.Join(workspacePath, "dyd", "traits")
	workspaceInfo, err := os.Lstat(workspaceTraitsPath)
	assert.Nil(err)
	assert.True(workspaceInfo.Mode()&os.ModeSymlink == os.ModeSymlink)
}

func TestRootBuildMaterializeVariantTraits_NoneConflictsWithExistingTrait(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "os"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "none"), "true")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=none")
	assert.NotNil(err)
	assert.Contains(err.Error(), "requires omitted trait")
}

func TestRootBuildMaterializeVariantTraits_RejectsNonConcreteKeywords(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "linux"), "true")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=inherit")
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid concrete root variant option")
}

func TestRootBuildMaterializeVariantTraits_OverwritesMismatchedTraitInWorkspace(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "os"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "dimensions", "os", "darwin"), "true")

	err := rootBuild_materializeVariantTraits(task.SERIAL_CONTEXT, rootPath, workspacePath, "os=darwin")
	assert.Nil(err)

	assert.Equal("darwin", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "os")))
	assert.Equal("linux", readTrimmedFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "os")))
}
