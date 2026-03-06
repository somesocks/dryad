package core

import (
	"dryad/diagnostics"
	dfilepath "dryad/internal/filepath"
	dydos "dryad/internal/os"

	"errors"
	stdos "os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeapMaterializeFile_RollsOverToReplicaOnEMLINK(t *testing.T) {
	assert := assert.New(t)

	diagnostics.Disable()
	t.Cleanup(diagnostics.Disable)

	sourcePath := filepath.Join(t.TempDir(), "source")
	destPath := filepath.Join(t.TempDir(), "dest")
	writeFileForTest(t, sourcePath, "same-content")

	err := stdos.Chmod(sourcePath, 0o511)
	assert.Nil(err)

	err = diagnostics.SetupFromConfig(diagnostics.Config{
		Version: 1,
		Seed:    1,
		Rules: []diagnostics.RuleConfig{
			{
				ID:   "inject-canonical-link-failure",
				Op:   "os.link",
				Key:  sourcePath,
				When: diagnostics.WhenConfig{Mode: "every_x", X: 1},
				Action: diagnostics.ActionConfig{
					Type:  "error",
					Error: "EMLINK",
				},
			},
		},
	})
	assert.Nil(err)

	err = heapMaterializeFile(sourcePath, destPath)
	assert.Nil(err)

	body, err := stdos.ReadFile(destPath)
	assert.Nil(err)
	assert.Equal("same-content", string(body))

	sourceInfo, err := dydos.Stat(sourcePath)
	assert.Nil(err)
	destInfo, err := dydos.Stat(destPath)
	assert.Nil(err)

	replicas, err := dfilepath.Glob(sourcePath + heapReplicaPathSeparator + "*")
	assert.Nil(err)
	assert.Len(replicas, 1)

	replicaInfo, err := dydos.Stat(replicas[0])
	assert.Nil(err)
	assert.Equal(sourceInfo.Mode().Perm(), destInfo.Mode().Perm())
	assert.Equal(fileInodeForTest(t, replicaInfo), fileInodeForTest(t, destInfo))
	assert.NotEqual(fileInodeForTest(t, sourceInfo), fileInodeForTest(t, destInfo))
}

func TestHeapMaterializeFile_PreservesPostErrorBehavior(t *testing.T) {
	assert := assert.New(t)

	diagnostics.Disable()
	t.Cleanup(diagnostics.Disable)

	sourcePath := filepath.Join(t.TempDir(), "source")
	destPath := filepath.Join(t.TempDir(), "dest")
	writeFileForTest(t, sourcePath, "same-content")

	err := diagnostics.SetupFromConfig(diagnostics.Config{
		Version: 1,
		Seed:    1,
		Rules: []diagnostics.RuleConfig{
			{
				ID:   "inject-post-link",
				Op:   "os.link",
				Key:  "*",
				When: diagnostics.WhenConfig{Mode: "every_x", X: 1},
				Action: diagnostics.ActionConfig{
					Type:  "error",
					Phase: "post",
					Error: "EMLINK",
				},
			},
		},
	})
	assert.Nil(err)

	err = heapMaterializeFile(sourcePath, destPath)
	assert.True(errors.Is(err, syscall.EMLINK))

	sourceInfo, err := dydos.Stat(sourcePath)
	assert.Nil(err)
	destInfo, err := dydos.Stat(destPath)
	assert.Nil(err)
	assert.Equal(fileInodeForTest(t, sourceInfo), fileInodeForTest(t, destInfo))

	replicas, err := dfilepath.Glob(sourcePath + heapReplicaPathSeparator + "*")
	assert.Nil(err)
	assert.Len(replicas, 0)
}

func fileInodeForTest(t *testing.T, info dydos.FileInfo) uint64 {
	t.Helper()

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("unexpected FileInfo.Sys type %T", info.Sys())
	}
	return stat.Ino
}
