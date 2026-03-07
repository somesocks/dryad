package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dryad/task"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
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

func testFingerprint(seed string) string {
	digest := blake2b.Sum256([]byte(seed))
	return fingerprintFormat(
		fingerprintVersionV2,
		fingerprintEncode(digest[:fingerprintDigestLen]),
	)
}

func TestHeapDerivationsAdd_WritesRegularFileInRootsNamespace(t *testing.T) {
	assert := assert.New(t)
	_, _, derivations := setupDerivationsForTest(t)

	sourceFingerprint := testFingerprint("source-fp")
	resultFingerprint := testFingerprint("result-fp")
	err, _ := derivations.Add(task.SERIAL_CONTEXT, sourceFingerprint, resultFingerprint)
	assert.Nil(err)

	derivationPath, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, sourceFingerprint)
	info, err := os.Lstat(derivationPath)
	assert.Nil(err)
	assert.True(info.Mode().IsRegular())

	bytes, err := os.ReadFile(derivationPath)
	assert.Nil(err)
	assert.Equal(resultFingerprint, strings.TrimSpace(string(bytes)))
}

func TestHeapDerivationExists_IgnoresLegacySymlinkEntries(t *testing.T) {
	assert := assert.New(t)
	_, heap, derivations := setupDerivationsForTest(t)

	resultFingerprint := testFingerprint("result-fp")
	targetStemPath, err := heapStemsFingerprintPath(filepath.Join(heap.BasePath, "stems"), resultFingerprint)
	assert.Nil(err)
	err = os.MkdirAll(filepath.Dir(targetStemPath), os.ModePerm)
	assert.Nil(err)

	sourceFingerprint := testFingerprint("source-fp")
	legacyPath, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, sourceFingerprint)
	err = os.Symlink(targetStemPath, legacyPath)
	assert.Nil(err)

	derivation := derivations.Derivation(sourceFingerprint)
	err, exists := derivation.Exists(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.False(exists)
}

func TestHeapDerivationResolve_FailsWhenResultStemMissing(t *testing.T) {
	assert := assert.New(t)
	_, _, derivations := setupDerivationsForTest(t)

	sourceFingerprint := testFingerprint("source-fp")
	derivationPath, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, sourceFingerprint)
	assert.Nil(err)
	writeFileForTest(t, derivationPath, testFingerprint("missing-result-fp"))

	derivation := derivations.Derivation(sourceFingerprint)
	err, _ = derivation.Resolve(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to resolve derivation")
}

func TestHeapDerivationResolve_ResolvesWhenResultStemExists(t *testing.T) {
	assert := assert.New(t)
	_, heap, derivations := setupDerivationsForTest(t)

	resultFingerprint := testFingerprint("result-fp")
	resultStemPath, err := heapStemsFingerprintPath(filepath.Join(heap.BasePath, "stems"), resultFingerprint)
	assert.Nil(err)
	err = os.MkdirAll(resultStemPath, os.ModePerm)
	assert.Nil(err)

	sourceFingerprint := testFingerprint("source-fp")
	derivationPath, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, sourceFingerprint)
	assert.Nil(err)
	writeFileForTest(t, derivationPath, resultFingerprint)

	derivation := derivations.Derivation(sourceFingerprint)
	err, safeRef := derivation.Resolve(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(resultStemPath, safeRef.Result.BasePath)
}

func TestGardenPruneSweepDerivations_RemovesStaleEntriesAndKeepsValidOnes(t *testing.T) {
	assert := assert.New(t)
	garden, heap, derivations := setupDerivationsForTest(t)

	validStem := testFingerprint("valid-stem-fp")
	validDerivation, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, testFingerprint("source-valid"))
	assert.Nil(err)
	staleDerivation, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, testFingerprint("source-stale"))
	assert.Nil(err)
	legacyDerivation, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, testFingerprint("source-legacy"))
	assert.Nil(err)

	validStemPath, err := heapStemsFingerprintPath(filepath.Join(heap.BasePath, "stems"), validStem)
	assert.Nil(err)
	err = os.MkdirAll(validStemPath, os.ModePerm)
	assert.Nil(err)
	writeFileForTest(t, validDerivation, validStem)
	writeFileForTest(t, staleDerivation, testFingerprint("missing-stem-fp"))

	err = os.Symlink(validStemPath, legacyDerivation)
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

	freshDerivation, err := heapDerivationsRootsFingerprintPath(derivations.BasePath, testFingerprint("source-fresh"))
	assert.Nil(err)
	writeFileForTest(t, freshDerivation, testFingerprint("missing-stem-fp"))

	req := gardenPruneRequest{
		Garden:   garden,
		Snapshot: time.Now().Add(-time.Second),
	}
	err, _ = gardenPrune_sweepDerivations(task.SERIAL_CONTEXT, req)
	assert.Nil(err)

	exists, err := fileExists(freshDerivation)
	assert.Nil(err)
	assert.True(exists)
}
