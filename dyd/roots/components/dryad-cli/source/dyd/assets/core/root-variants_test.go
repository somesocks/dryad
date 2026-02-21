package core

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dryad/task"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func writeFileForTest(t *testing.T, path string, content string) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(path, []byte(content), os.ModePerm)
	assert.Nil(t, err)
}

func captureCoreLogsForTest(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	var out bytes.Buffer
	oldLogger := zlog.Logger
	oldLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	zlog.Logger = zerolog.New(&out)

	return &out, func() {
		zlog.Logger = oldLogger
		zerolog.SetGlobalLevel(oldLevel)
	}
}

func TestRootVariantsDimensionsLoad_Basic(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "darwin"), "false")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "amd64"), "true")
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "arch", "arm64"), "true\n")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, dimensions := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(
		[]VariantDimension{
			{
				Name: "arch",
				Options: []VariantDimensionOption{
					{Name: "amd64", Enabled: true},
					{Name: "arm64", Enabled: true},
				},
			},
			{
				Name: "os",
				Options: []VariantDimensionOption{
					{Name: "darwin", Enabled: false},
					{Name: "linux", Enabled: true},
				},
			},
		},
		dimensions,
	)
}

func TestRootVariantsDimensionsLoad_MissingVariants(t *testing.T) {
	assert := assert.New(t)

	root := SafeRootReference{
		BasePath: t.TempDir(),
	}

	err, dimensions := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal([]VariantDimension{}, dimensions)
}

func TestRootVariantsDimensionsLoad_RejectsReservedOption(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "inherit"), "true")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, _ := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "reserved variant option")
}

func TestRootVariantsDimensionsLoad_RejectsInvalidOptionValue(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux"), "yes")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, _ := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "true or false")
}

func TestRootVariantsDimensionsLoad_RejectsInvalidNames(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	writeFileForTest(t, filepath.Join(rootPath, "dyd", "traits", "variants", "bad name", "linux"), "true")

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, _ := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid variant dimension name")
}

func TestRootVariantsDimensionsLoad_WarnsOnWhitespaceInOptionFile(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	optionPath := filepath.Join(rootPath, "dyd", "traits", "variants", "os", "linux")
	writeFileForTest(t, optionPath, "true\n")

	logOutput, restoreLogs := captureCoreLogsForTest(t)
	defer restoreLogs()

	root := SafeRootReference{
		BasePath: rootPath,
	}

	err, dimensions := root.VariantDimensions(task.SERIAL_CONTEXT)
	assert.Nil(err)
	assert.Equal(
		[]VariantDimension{
			{
				Name: "os",
				Options: []VariantDimensionOption{
					{Name: "linux", Enabled: true},
				},
			},
		},
		dimensions,
	)

	logs := logOutput.String()
	assert.Contains(logs, "malformed variant option file")
	assert.Contains(logs, strings.ReplaceAll(optionPath, "\\", "\\\\"))
}
