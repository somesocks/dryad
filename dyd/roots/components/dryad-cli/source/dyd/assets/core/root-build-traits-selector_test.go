package core

import (
	"dryad/internal/filepath"
	"os"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectTraitsPath_PlainTraitsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "ok")

	err, traitsPath := rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "traits"), traitsPath)
}

func TestRootBuildSelectTraitsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=linux", "name"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=darwin", "name"), "darwin")

	err, traitsPath := rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "traits~os=linux"), traitsPath)
}

func TestRootBuildSelectTraitsPath_NoneAnyAndOptionListsAreSupported(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=linux,none", "name"), "none-or-linux")

	err, traitsPath := rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "traits~os=linux,none"), traitsPath)

	err, traitsPath = rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "traits~os=linux,none"), traitsPath)
}

func TestRootBuildSelectTraitsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=inherit", "name"), "bad")

	err, _ := rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for traits variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=host", "name"), "bad")

	err, _ = rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for traits variant selectors")
}

func TestRootBuildSelectTraitsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "name"), "default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=linux", "name"), "linux")

	err, _ := rootBuild_selectTraitsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/traits selectors")
}

func TestRootBuildStage0_MaterializesOnlySelectedTraits(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=linux", "name"), "linux-name")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=darwin", "name"), "darwin-name")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=darwin", "darwin-only"), "darwin-only")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	assert.Equal("linux-name", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "name")))
	assert.Equal("linux", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "os")))

	darwinTraitExists, err := fileExists(filepath.Join(workspacePath, "dyd", "traits", "darwin-only"))
	assert.Nil(err)
	assert.False(darwinTraitExists)
}

func TestRootBuildStage0_NoMatchingTraitsStillMaterializesVariantTraits(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=darwin", "name"), "darwin-name")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	assert.Equal("linux", readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "traits", "os")))

	nameExists, err := fileExists(filepath.Join(workspacePath, "dyd", "traits", "name"))
	assert.Nil(err)
	assert.False(nameExists)
}

func TestRootBuildStage0_SelectedTraitsAreMaterializedAsDirectory(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits~os=linux", "name"), "linux-name")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	info, err := os.Lstat(filepath.Join(workspacePath, "dyd", "traits"))
	assert.Nil(err)
	assert.True(info.IsDir())
	assert.False(info.Mode()&os.ModeSymlink == os.ModeSymlink)
}
