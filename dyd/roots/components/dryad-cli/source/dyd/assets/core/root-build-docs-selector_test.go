package core

import (
	"os"
	"path/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectDocsPath_PlainDocsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs", "about.md"), "ok")

	err, docsPath := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs"), docsPath)
}

func TestRootBuildSelectDocsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=linux", "about.md"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=darwin", "about.md"), "darwin")

	err, docsPath := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs~os=linux"), docsPath)
}

func TestRootBuildSelectDocsPath_OmittedSelectorDimensionsAreImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=linux", "about.md"), "linux-any-arch")

	err, docsPath := rootBuild_selectDocsPathForTest(
		task.SERIAL_CONTEXT,
		rootPath,
		"arch=amd64+os=linux",
	)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs~os=linux"), docsPath)
}

func TestRootBuildSelectDocsPath_NoMatchesIsAllowed(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=darwin", "about.md"), "darwin")

	err, docsPath := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal("", docsPath)
}

func TestRootBuildSelectDocsPath_NoneAnyAndOptionListsAreSupported(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=linux,none", "about.md"), "none-or-linux")

	err, docsPath := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs~os=linux,none"), docsPath)

	err, docsPath = rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs~os=linux,none"), docsPath)
}

func TestRootBuildSelectDocsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=inherit", "about.md"), "bad")

	err, _ := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for docs variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=host", "about.md"), "bad")

	err, _ = rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for docs variant selectors")
}

func TestRootBuildSelectDocsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs", "about.md"), "default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=linux", "about.md"), "linux")

	err, _ := rootBuild_selectDocsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/docs selectors")
}

func TestRootBuildStage0_LinksOnlySelectedDocs(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=linux", "about.md"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=darwin", "about.md"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	workspaceDocsPath := filepath.Join(workspacePath, "dyd", "docs")
	workspaceDocsInfo, err := os.Lstat(workspaceDocsPath)
	assert.Nil(err)
	assert.True(workspaceDocsInfo.Mode()&os.ModeSymlink == os.ModeSymlink)

	selectedDocsPath, err := os.Readlink(workspaceDocsPath)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "docs~os=linux"), selectedDocsPath)
}

func TestRootBuildStage0_NoMatchingDocsLeavesDocsPathAbsent(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "docs~os=darwin", "about.md"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	exists, err := fileExists(filepath.Join(workspacePath, "dyd", "docs"))
	assert.Nil(err)
	assert.False(exists)
}
