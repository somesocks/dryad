package core

import (
	"dryad/internal/filepath"
	"os"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectCommandsPath_PlainCommandsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands", "dyd-root-build"), "ok")

	err, commandsPath := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands"), commandsPath)
}

func TestRootBuildSelectCommandsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=linux", "dyd-root-build"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=darwin", "dyd-root-build"), "darwin")

	err, commandsPath := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands~os=linux"), commandsPath)
}

func TestRootBuildSelectCommandsPath_OmittedSelectorDimensionsAreImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=linux", "dyd-root-build"), "linux-any-arch")

	err, commandsPath := rootBuild_selectCommandsPathForTest(
		task.SERIAL_CONTEXT,
		rootPath,
		"arch=amd64+os=linux",
	)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands~os=linux"), commandsPath)
}

func TestRootBuildSelectCommandsPath_NoMatchesIsAllowed(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=darwin", "dyd-root-build"), "darwin")

	err, commandsPath := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal("", commandsPath)
}

func TestRootBuildSelectCommandsPath_NoneAnyAndOptionListsAreSupported(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=linux,none", "dyd-root-build"), "none-or-linux")

	err, commandsPath := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands~os=linux,none"), commandsPath)

	err, commandsPath = rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands~os=linux,none"), commandsPath)
}

func TestRootBuildSelectCommandsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=inherit", "dyd-root-build"), "bad")

	err, _ := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for commands variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=host", "dyd-root-build"), "bad")

	err, _ = rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for commands variant selectors")
}

func TestRootBuildSelectCommandsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands", "dyd-root-build"), "default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=linux", "dyd-root-build"), "linux")

	err, _ := rootBuild_selectCommandsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/commands selectors")
}

func TestRootBuildStage0_LinksOnlySelectedCommands(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=linux", "dyd-root-build"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=darwin", "dyd-root-build"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	workspaceCommandsPath := filepath.Join(workspacePath, "dyd", "commands")
	workspaceCommandsInfo, err := os.Lstat(workspaceCommandsPath)
	assert.Nil(err)
	assert.True(workspaceCommandsInfo.Mode()&os.ModeSymlink == os.ModeSymlink)

	selectedCommandsPath, err := os.Readlink(workspaceCommandsPath)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "commands~os=linux"), selectedCommandsPath)
}

func TestRootBuildStage0_NoMatchingCommandsLeavesCommandsPathAbsent(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "commands~os=darwin", "dyd-root-build"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	exists, err := fileExists(filepath.Join(workspacePath, "dyd", "commands"))
	assert.Nil(err)
	assert.False(exists)
}
