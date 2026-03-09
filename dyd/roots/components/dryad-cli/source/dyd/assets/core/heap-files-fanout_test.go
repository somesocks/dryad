package core

import (
	"dryad/internal/filepath"
	"os"
	"strconv"
	"testing"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func setupHeapFilesForTest(t *testing.T) (*SafeGardenReference, *SafeHeapFilesReference) {
	t.Helper()

	gardenPath := t.TempDir()
	heapPath := filepath.Join(gardenPath, "dyd", "heap")
	err := os.MkdirAll(heapPath, os.ModePerm)
	assert.Nil(t, err)

	garden := &SafeGardenReference{BasePath: gardenPath}
	heap := &SafeHeapReference{
		BasePath: heapPath,
		Garden:   garden,
	}

	err, heapFiles := heap.Files().Resolve(task.SERIAL_CONTEXT)
	assert.Nil(t, err)

	return garden, heapFiles
}

func writeHeapFilesDepthForTest(t *testing.T, garden *SafeGardenReference, depth int) {
	t.Helper()
	writeFileForTest(t, shedHeapFilesDepthPath(safeShedReference(garden)), strconv.Itoa(depth))
}

func writeSourceFileForHeapFilesTest(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o644)
	assert.Nil(t, err)
	return path
}

func TestHeapFilesAddFile_UsesFlatLayoutWhenDepthZero(t *testing.T) {
	assert := assert.New(t)
	garden, heapFiles := setupHeapFilesForTest(t)
	writeHeapFilesDepthForTest(t, garden, 0)

	sourcePath := writeSourceFileForHeapFilesTest(t, t.TempDir(), "source.txt", "hello")
	ctx := task.NewContext(1)
	err, fingerprint := heapFiles.AddFile(ctx, sourcePath)
	assert.Nil(err)

	err, version, encoded := fingerprintParse(fingerprint)
	assert.Nil(err)
	assert.Equal(fingerprintVersionV2, version)

	err, canonicalPath := heapFilesFingerprintPath(ctx, garden, heapFiles.BasePath, fingerprint)
	assert.Nil(err)
	assert.Equal(filepath.Join(heapFiles.BasePath, version, encoded), canonicalPath)

	_, err = os.Stat(canonicalPath)
	assert.Nil(err)
}

func TestHeapFilesAddFile_UsesFanoutLayoutWhenDepthTwo(t *testing.T) {
	assert := assert.New(t)
	garden, heapFiles := setupHeapFilesForTest(t)
	writeHeapFilesDepthForTest(t, garden, 2)

	sourcePath := writeSourceFileForHeapFilesTest(t, t.TempDir(), "source.txt", "hello")
	ctx := task.NewContext(1)
	err, fingerprint := heapFiles.AddFile(ctx, sourcePath)
	assert.Nil(err)

	err, version, encoded := fingerprintParse(fingerprint)
	assert.Nil(err)

	err, canonicalPath := heapFilesFingerprintPath(ctx, garden, heapFiles.BasePath, fingerprint)
	assert.Nil(err)
	assert.Equal(
		filepath.Join(heapFiles.BasePath, version, encoded[:2], encoded[2:4], encoded[4:]),
		canonicalPath,
	)

	_, err = os.Stat(canonicalPath)
	assert.Nil(err)
}

func TestHeapFilesAddFile_UsesUpdatedDepthWithNewContext(t *testing.T) {
	assert := assert.New(t)
	garden, heapFiles := setupHeapFilesForTest(t)
	writeHeapFilesDepthForTest(t, garden, 0)

	sourceA := writeSourceFileForHeapFilesTest(t, t.TempDir(), "a.txt", "alpha")
	ctxA := task.NewContext(1)
	err, fingerprintA := heapFiles.AddFile(ctxA, sourceA)
	assert.Nil(err)

	err, pathA := heapFilesFingerprintPath(ctxA, garden, heapFiles.BasePath, fingerprintA)
	assert.Nil(err)
	_, err = os.Stat(pathA)
	assert.Nil(err)

	writeHeapFilesDepthForTest(t, garden, 2)

	sourceB := writeSourceFileForHeapFilesTest(t, t.TempDir(), "b.txt", "beta")
	ctxB := task.NewContext(1)
	err, fingerprintB := heapFiles.AddFile(ctxB, sourceB)
	assert.Nil(err)

	err, pathB := heapFilesFingerprintPath(ctxB, garden, heapFiles.BasePath, fingerprintB)
	assert.Nil(err)
	_, err = os.Stat(pathB)
	assert.Nil(err)

	err, version, encoded := fingerprintParse(fingerprintB)
	assert.Nil(err)

	assert.Equal(filepath.Join(heapFiles.BasePath, version, encoded[:2], encoded[2:4], encoded[4:]), pathB)
	_, err = os.Stat(filepath.Join(heapFiles.BasePath, version, encoded))
	assert.True(os.IsNotExist(err))
}

func TestHeapFilesFingerprintPath_MemoizesDepthWithinContext(t *testing.T) {
	assert := assert.New(t)
	garden, heapFiles := setupHeapFilesForTest(t)
	fingerprint := testFingerprint("memoized-depth")

	writeHeapFilesDepthForTest(t, garden, 0)
	ctxA := task.NewContext(1)
	err, flatPath := heapFilesFingerprintPath(ctxA, garden, heapFiles.BasePath, fingerprint)
	assert.Nil(err)

	writeHeapFilesDepthForTest(t, garden, 2)

	err, stillFlatPath := heapFilesFingerprintPath(ctxA, garden, heapFiles.BasePath, fingerprint)
	assert.Nil(err)
	assert.Equal(flatPath, stillFlatPath)

	ctxB := task.NewContext(1)
	err, fanoutPath := heapFilesFingerprintPath(ctxB, garden, heapFiles.BasePath, fingerprint)
	assert.Nil(err)
	assert.NotEqual(flatPath, fanoutPath)
}
