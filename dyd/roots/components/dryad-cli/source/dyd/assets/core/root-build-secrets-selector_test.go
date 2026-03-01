package core

import (
	"os"
	"path/filepath"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootBuildSelectSecretsPath_PlainSecretsMatchesImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets", "main"), "ok")

	err, secretsPath := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets"), secretsPath)
}

func TestRootBuildSelectSecretsPath_ConditionalSelectorMatchesConcreteVariant(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=linux", "main"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=darwin", "main"), "darwin")

	err, secretsPath := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets~os=linux"), secretsPath)
}

func TestRootBuildSelectSecretsPath_OmittedSelectorDimensionsAreImplicitAny(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=linux", "main"), "linux-any-arch")

	err, secretsPath := rootBuild_selectSecretsPathForTest(
		task.SERIAL_CONTEXT,
		rootPath,
		"arch=amd64+os=linux",
	)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets~os=linux"), secretsPath)
}

func TestRootBuildSelectSecretsPath_NoMatchesIsAllowed(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=darwin", "main"), "darwin")

	err, secretsPath := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal("", secretsPath)
}

func TestRootBuildSelectSecretsPath_NoneAnyAndOptionListsAreSupported(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "none"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=linux,none", "main"), "none-or-linux")

	err, secretsPath := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets~os=linux,none"), secretsPath)

	err, secretsPath = rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets~os=linux,none"), secretsPath)
}

func TestRootBuildSelectSecretsPath_InheritAndHostAreRejected(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=inherit", "main"), "bad")

	err, _ := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "inherit option is not supported for secrets variant selectors")

	rootPath = t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=host", "main"), "bad")

	err, _ = rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "host option is not supported for secrets variant selectors")
}

func TestRootBuildSelectSecretsPath_MultipleMatchesFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets", "main"), "default")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=linux", "main"), "linux")

	err, _ := rootBuild_selectSecretsPathForTest(task.SERIAL_CONTEXT, rootPath, "os=linux")
	assert.NotNil(err)
	assert.Contains(err.Error(), "multiple matching dyd/secrets selectors")
}

func TestRootBuildStage0_LinksOnlySelectedSecrets(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=linux", "main"), "linux")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=darwin", "main"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	workspaceSecretsPath := filepath.Join(workspacePath, "dyd", "secrets")
	workspaceSecretsInfo, err := os.Lstat(workspaceSecretsPath)
	assert.Nil(err)
	assert.True(workspaceSecretsInfo.Mode()&os.ModeSymlink == os.ModeSymlink)

	selectedSecretsPath, err := os.Readlink(workspaceSecretsPath)
	assert.Nil(err)
	assert.Equal(filepath.Join(rootPath, "dyd", "secrets~os=linux"), selectedSecretsPath)
}

func TestRootBuildStage0_NoMatchingSecretsLeavesSecretsPathAbsent(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	workspacePath := t.TempDir()

	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "secrets~os=darwin", "main"), "darwin")

	err, _ := rootBuild_stage0(task.SERIAL_CONTEXT, rootBuild_stage0_request{
		RootPath:          rootPath,
		WorkspacePath:     workspacePath,
		VariantDescriptor: "os=linux",
	})
	assert.Nil(err)

	exists, err := fileExists(filepath.Join(workspacePath, "dyd", "secrets"))
	assert.Nil(err)
	assert.False(exists)
}
