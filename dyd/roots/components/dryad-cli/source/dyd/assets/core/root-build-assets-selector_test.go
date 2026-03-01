package core

import (
	"os"
	"path/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectAssetsPath_PlainAssetsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets", "main"), "ok")

	err, assetsPath := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets"), assetsPath)
}

func TestRootBuildSelectAssetsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=linux", "main"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=darwin", "main"), "darwin")

	err, assetsPath := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets~os=linux"), assetsPath)
}

func TestRootBuildSelectAssetsPath_OmittedSelectorDimensionsAreImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=linux", "main"), "linux-any-arch")

	err, assetsPath := rootBuild_selectAssetsPath(
		task.SERIAL_CONTEXT,
		rootPath,
		"arch=amd64+os=linux",
	)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets~os=linux"), assetsPath)
}

func TestRootBuildSelectAssetsPath_NoMatchesIsAllowed(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=darwin", "main"), "darwin")

	err, assetsPath := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal("", assetsPath)
}

func TestRootBuildSelectAssetsPath_NoneAnyAndOptionListsAreSupported(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=linux,none", "main"), "none-or-linux")

	err, assetsPath := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets~os=linux,none"), assetsPath)

	err, assetsPath = rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets~os=linux,none"), assetsPath)
}

func TestRootBuildSelectAssetsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=inherit", "main"), "bad")

	err, _ := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for assets variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=host", "main"), "bad")

	err, _ = rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for assets variant selectors")
}

func TestRootBuildSelectAssetsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets", "main"), "default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=linux", "main"), "linux")

	err, _ := rootBuild_selectAssetsPath(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/assets selectors")
}

func TestRootBuildStage0_LinksOnlySelectedAssets(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=linux", "main"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=darwin", "main"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	workspaceAssetsPath := filepath.Join(workspacePath, "dyd", "assets")
	workspaceAssetsInfo, err := os.Lstat(workspaceAssetsPath)
	assert.Nil(err)
	assert.True(workspaceAssetsInfo.Mode()&os.ModeSymlink == os.ModeSymlink)

	selectedAssetsPath, err := os.Readlink(workspaceAssetsPath)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "assets~os=linux"), selectedAssetsPath)
}

func TestRootBuildStage0_NoMatchingAssetsLeavesAssetsPathAbsent(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "assets~os=darwin", "main"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	exists, err := fileExists(filepath.Join(workspacePath, "dyd", "assets"))
	assert.Nil(err)
	assert.False(exists)
}
