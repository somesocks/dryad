package core

import (
	"dryad/internal/filepath"
	"os"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectRequirementsPath_PlainRequirementsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements", "dep"), "root:../../../dep")

	err, requirementsPath := rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "requirements"), requirementsPath)
}

func TestRootBuildSelectRequirementsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=linux", "dep"), "root:../../../dep-linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=darwin", "dep"), "root:../../../dep-darwin")

	err, requirementsPath := rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "requirements~os=linux"), requirementsPath)
}

func TestRootBuildSelectRequirementsPath_NoMatchesIsAllowed(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=darwin", "dep"), "root:../../../dep-darwin")

	err, requirementsPath := rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal("", requirementsPath)
}

func TestRootBuildSelectRequirementsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=inherit", "dep"), "root:../../../dep")

	err, _ := rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for requirements variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=host", "dep"), "root:../../../dep")

	err, _ = rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for requirements variant selectors")
}

func TestRootBuildSelectRequirementsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements", "dep"), "root:../../../dep-default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=linux", "dep"), "root:../../../dep-linux")

	err, _ := rootBuild_selectRequirementsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/requirements selectors")
}

func TestRootBuildStage0_LinksOnlySelectedRequirementsToTempPath(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=linux", "dep"), "root:../../../dep-linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=darwin", "dep"), "root:../../../dep-darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	workspaceRequirementsPath := filepath.Join(workspacePath, "dyd", "~requirements")
	workspaceRequirementsInfo, err := os.Lstat(workspaceRequirementsPath)
	assert.Nil(err)
	assert.True(workspaceRequirementsInfo.Mode()&os.ModeSymlink == os.ModeSymlink)

	selectedRequirementsPath, err := os.Readlink(workspaceRequirementsPath)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "requirements~os=linux"), selectedRequirementsPath)
}

func TestRootBuildStage0_NoMatchingRequirementsLeavesTempRequirementsPathAbsent(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "requirements~os=darwin", "dep"), "root:../../../dep-darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	exists, err := fileExists(filepath.Join(workspacePath, "dyd", "~requirements"))
	assert.Nil(err)
	assert.False(exists)
}
