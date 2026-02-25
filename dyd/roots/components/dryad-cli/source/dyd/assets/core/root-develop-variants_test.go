package core

import (
	"path/filepath"
	"runtime"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func TestRootDevelopResolveVariant_DefaultAmbiguousFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")

	root := &SafeRootReference{BasePath: rootPath}
	err, _ := rootDevelop_resolveVariant(task.SERIAL_CONTEXT, root, "")
	assert.NotNil(err)
	assert.Contains(err.Error(), "ambiguous root develop variant selector")
}

func TestRootDevelopResolveVariant_DefaultUniquePasses(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "false")

	root := &SafeRootReference{BasePath: rootPath}
	err, descriptor := rootDevelop_resolveVariant(task.SERIAL_CONTEXT, root, "")
	assert.Nil(err)
	assert.Equal("os=linux", descriptor)
}

func TestRootDevelopResolveVariant_PartialSelectorAmbiguousFails(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")

	root := &SafeRootReference{BasePath: rootPath}
	err, _ := rootDevelop_resolveVariant(task.SERIAL_CONTEXT, root, "arch=amd64")
	assert.NotNil(err)
	assert.Contains(err.Error(), "ambiguous root develop variant selector")
}

func TestRootDevelopResolveVariant_PartialSelectorUniquePasses(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "true")

	root := &SafeRootReference{BasePath: rootPath}
	err, descriptor := rootDevelop_resolveVariant(task.SERIAL_CONTEXT, root, "arch=amd64+os=linux")
	assert.Nil(err)
	assert.Equal("arch=amd64+os=linux", descriptor)
}

func TestRootDevelopResolveVariant_HostSelectorUniquePasses(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", runtime.GOOS), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "other"), "true")

	root := &SafeRootReference{BasePath: rootPath}
	err, descriptor := rootDevelop_resolveVariant(task.SERIAL_CONTEXT, root, "os=host")
	assert.Nil(err)
	assert.Equal("os="+runtime.GOOS, descriptor)
}
