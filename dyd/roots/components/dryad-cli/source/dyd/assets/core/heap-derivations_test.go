package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dryad/task"
	"github.com/stretchr/testify/assert"
)

func setupDerivationsForTest(t *testing.T) (*SafeGardenReference, *SafeHeapReference, *SafeHeapDerivationsReference) {
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

	err, derivations := heap.Derivations().Resolve(task.SERIAL_CONTEXT)
	assert.Nil(t, err)

	return garden, heap, derivations
}

func TestHeapDerivationsAdd_WritesRegularFileInRootsNamespace(t *testing.T) {
	assert := assert.New(t)
	_, _, derivations := setupDerivationsForTest(t)

	err, _ := derivations.Add(task.SERIAL_CONTEXT, "source-fp", "result-fp")
	assert.Nil(err)

	derivationPath := filepath.Join(derivations.BasePath, "roots", "source-fp")
	info, err := os.Lstat(derivationPath)
	assert.Nil(err)
	assert.True(info.Mode().IsRegular())

	bytes, err := os.ReadFile(derivationPath)
	assert.Nil(err)
	assert.Equal("result-fp", strings.TrimSpace(string(bytes)))
}

func TestHeapDerivationExists_IgnoresLegacySymlinkEntries(t *testing.T) {
	assert := assert.New(t)
	_, heap, derivations := setupDerivationsForTest(t)

	targetStemPath := filepath.Join(heap.BasePath, "stems", "result-fp")
	err := os.MkdirAll(filepath.Dir(targetStemPath), os.ModePerm)
	assert.Nil(err)

	legacyPath := filepath.Join(derivations.BasePath, "roots", "source-fp")
	err = os.Symlink(targetStemPath, legacyPath)
	assert.Nil(err)

	err, exists := derivations.Derivation("source-fp").Exists(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.False(exists)
}

func TestHeapDerivationResolve_FailsWhenResultStemMissing(t *testing.T) {
	assert := assert.New(t)
	_, _, derivations := setupDerivationsForTest(t)

	derivationPath := filepath.Join(derivations.BasePath, "roots", "source-fp")
	writeFileForTest(t, derivationPath, "missing-result-fp")

	err, _ := derivations.Derivation("source-fp").Resolve(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to resolve derivation")
}

func TestHeapDerivationResolve_ResolvesWhenResultStemExists(t *testing.T) {
	assert := assert.New(t)
	_, heap, derivations := setupDerivationsForTest(t)

	resultFingerprint := "result-fp"
	err := os.MkdirAll(filepath.Join(heap.BasePath, "stems", resultFingerprint), os.ModePerm)
	assert.Nil(err)

	derivationPath := filepath.Join(derivations.BasePath, "roots", "source-fp")
	writeFileForTest(t, derivationPath, resultFingerprint)

	err, safeRef := derivations.Derivation("source-fp").Resolve(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(resultFingerprint, filepath.Base(safeRef.Result.BasePath))
}

func TestGardenPruneSweepDerivations_RemovesStaleEntriesAndKeepsValidOnes(t *testing.T) {
	assert := assert.New(t)
	garden, heap, derivations := setupDerivationsForTest(t)

	validStem := "valid-stem-fp"
	validDerivation := filepath.Join(derivations.BasePath, "roots", "source-valid")
	staleDerivation := filepath.Join(derivations.BasePath, "roots", "source-stale")
	legacyDerivation := filepath.Join(derivations.BasePath, "roots", "source-legacy")

	err := os.MkdirAll(filepath.Join(heap.BasePath, "stems", validStem), os.ModePerm)
	assert.Nil(err)
	writeFileForTest(t, validDerivation, validStem)
	writeFileForTest(t, staleDerivation, "missing-stem-fp")

	err = os.Symlink(filepath.Join(heap.BasePath, "stems", validStem), legacyDerivation)
	assert.Nil(err)

	req := gardenPruneRequest{
		Garden:   garden,
		Snapshot: time.Now().Add(time.Hour),
	}
	err, _ = gardenPrune_sweepDerivations(task.SERIAL_CONTEXT, req)
	assert.Nil(err)

	validExists, err := fileExists(validDerivation)
	assert.Nil(err)
	assert.True(validExists)

	staleExists, err := fileExists(staleDerivation)
	assert.Nil(err)
	assert.False(staleExists)

	legacyExists, err := fileExists(legacyDerivation)
	assert.Nil(err)
	assert.False(legacyExists)
}

func TestGardenPruneSweepDerivations_SkipsFreshEntriesAfterSnapshot(t *testing.T) {
	assert := assert.New(t)
	garden, _, derivations := setupDerivationsForTest(t)

	freshDerivation := filepath.Join(derivations.BasePath, "roots", "source-fresh")
	writeFileForTest(t, freshDerivation, "missing-stem-fp")

	req := gardenPruneRequest{
		Garden:   garden,
		Snapshot: time.Now().Add(-time.Second),
	}
	err, _ := gardenPrune_sweepDerivations(task.SERIAL_CONTEXT, req)
	assert.Nil(err)

	exists, err := fileExists(freshDerivation)
	assert.Nil(err)
	assert.True(exists)
}
