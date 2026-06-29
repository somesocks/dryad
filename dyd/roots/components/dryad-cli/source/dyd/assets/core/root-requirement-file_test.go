package core

import (
	"archive/tar"
	"archive/zip"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"encoding/base64"
	stdos "os"
	stdfilepath "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeWritableForCleanupForTest(t *testing.T, path string) {
	t.Helper()
	t.Cleanup(func() {
		_ = stdfilepath.WalkDir(path, func(path string, entry stdos.DirEntry, err error) error {
			if err != nil || entry.Type()&stdos.ModeSymlink == stdos.ModeSymlink {
				return nil
			}
			if entry.IsDir() {
				_ = stdos.Chmod(path, 0o755)
				return nil
			}
			_ = stdos.Chmod(path, 0o644)
			return nil
		})
	})
}

func TestRootRequirementFileTargetNormalize(t *testing.T) {
	assert := assert.New(t)

	err, target := RootRequirementFileTargetNormalize("file:../foo.txt")
	assert.Nil(err)
	assert.Equal("file:../foo.txt", target)

	err, target = RootRequirementFileTargetNormalize("file:../foo.txt?into=dyd/assets")
	assert.Nil(err)
	assert.Equal("file:../foo.txt", target)

	err, target = RootRequirementFileTargetNormalize("file:../foo.txt?as=dyd/secrets/.env&unpack=true&optional=true&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Nil(err)
	assert.Equal("file:../foo.txt?as=dyd/secrets/.env&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa&optional=true&unpack=true", target)

	err, _ = RootRequirementFileTargetNormalize("file:/abs/foo.txt")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?target=assets")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?as=dyd/assets/foo.txt&into=dyd/assets")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?as=/dyd/assets/foo.txt")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?as=dyd/assets/../../outside.txt")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?as=dyd/assets~os=linux/foo.txt")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?as=dyd/commands/foo.txt")
	assert.NotNil(err)

	err, _ = RootRequirementFileTargetNormalize("file:../foo.txt?optional=maybe")
	assert.NotNil(err)
}

func TestRootRequirementFileBuildStem_DirectoryHonorsIgnoreAndSymlinks(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "source")
	externalPath := filepath.Join(t.TempDir(), "external.txt")
	writeFileForTest(t, filepath.Join(sourcePath, ".dyd-ignore"), "ignored.txt\nignored-dir\n")
	writeFileForTest(t, filepath.Join(sourcePath, "keep.txt"), "keep")
	writeFileForTest(t, filepath.Join(sourcePath, "ignored.txt"), "ignored")
	writeFileForTest(t, filepath.Join(sourcePath, "ignored-dir", "value.txt"), "ignored")
	writeFileForTest(t, externalPath, "external")
	assert.Nil(os.Symlink("keep.txt", filepath.Join(sourcePath, "internal-link")))
	assert.Nil(os.Symlink(externalPath, filepath.Join(sourcePath, "external-link")))

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourcePath:    sourcePath,
		DestinationAs: "dyd/assets/vendor",
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.FileExists(filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "keep.txt"))
	assert.NoFileExists(filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "ignored.txt"))
	assert.NoDirExists(filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "ignored-dir"))
	linkInfo, err := os.Lstat(filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "internal-link"))
	assert.Nil(err)
	assert.True(linkInfo.Mode()&os.ModeSymlink == os.ModeSymlink)
	assert.Equal("external", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "external-link")))
}

func TestRootRequirementFileBuildStem_FileAsSecrets(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), ".env")
	writeFileForTest(t, sourcePath, "SECRET=1")

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourcePath:    sourcePath,
		DestinationAs: "dyd/secrets/runtime.env",
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("SECRET=1", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "secrets", "runtime.env")))
}

func TestRootRequirementFileBuildStem_FileIntoUsesSourceName(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "value.txt")
	writeFileForTest(t, sourcePath, "value")

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      sourcePath,
		DestinationInto: "dyd/assets/config",
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("value", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "config", "value.txt")))
}

func TestRootRequirementFileBuildStem_DefaultTargetAssets(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "value.txt")
	writeFileForTest(t, sourcePath, "value")

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:     &SafeGardenReference{BasePath: gardenPath},
		SourcePath: sourcePath,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("value", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "value.txt")))
}

func TestRootRequirementFileBuildStem_MissingRequiredSourceFails(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "missing.txt")

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:     &SafeGardenReference{BasePath: gardenPath},
		SourcePath: sourcePath,
	})
	assert.NotNil(err)
	assert.Nil(stem)
}

func TestRootRequirementFileBuildStem_MissingOptionalFileBuildsEmptyStem(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "missing.txt")

	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourcePath:    sourcePath,
		DestinationAs: "dyd/assets/missing.txt",
		Optional:      true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("stem", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "type")))
	assert.FileExists(filepath.Join(stem.BasePath, "dyd", "fingerprint"))
	assert.NoFileExists(filepath.Join(stem.BasePath, "dyd", "assets", "missing.txt"))
}

func TestRootRequirementFileBuildStem_MissingOptionalArchiveBuildsEmptyStem(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	sourcePath := filepath.Join(t.TempDir(), "missing.tar")

	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      sourcePath,
		DestinationInto: "dyd/assets/vendor",
		Optional:        true,
		Unpack:          true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.FileExists(filepath.Join(stem.BasePath, "dyd", "fingerprint"))
	assert.NoDirExists(filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "missing"))
}

func TestRootRequirementFileBuildStem_UnpackIntoUsesArchiveName(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "pkg.tar")
	archiveFile, err := stdos.Create(archivePath)
	assert.Nil(err)
	tarWriter := tar.NewWriter(archiveFile)
	contents := "packed"
	assert.Nil(tarWriter.WriteHeader(&tar.Header{
		Name: "contents/value.txt",
		Mode: 0o644,
		Size: int64(len(contents)),
	}))
	_, err = tarWriter.Write([]byte(contents))
	assert.Nil(err)
	assert.Nil(tarWriter.Close())
	assert.Nil(archiveFile.Close())

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      archivePath,
		DestinationInto: "dyd/assets/vendor",
		Unpack:          true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("packed", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "pkg", "contents", "value.txt")))
}

func TestRootRequirementFileBuildStem_UnpackZipIntoUsesArchiveName(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "pkg.zip")
	archiveFile, err := stdos.Create(archivePath)
	assert.Nil(err)
	zipWriter := zip.NewWriter(archiveFile)
	contents := "packed"
	entry, err := zipWriter.Create("contents/value.txt")
	assert.Nil(err)
	_, err = entry.Write([]byte(contents))
	assert.Nil(err)
	assert.Nil(zipWriter.Close())
	assert.Nil(archiveFile.Close())

	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      archivePath,
		DestinationInto: "dyd/assets/vendor",
		Unpack:          true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("packed", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "pkg", "contents", "value.txt")))
}

func TestRootRequirementFileBuildStem_UnpackTarBz2IntoUsesArchiveName(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "pkg.tar.bz2")
	archiveBytes, err := base64.StdEncoding.DecodeString("QlpoOTFBWSZTWYISVgoAAHd9hMEAAERAAf+AAAFuDd9AAACACCAAdBpNGk0NANANGnqCSgQAAANA4+jQ5yEGkAEiyBbeNeVMSQNYNmbrF+JILN5GGMErTawNSSMOMX6hHGmVaCBjymnO2dJpITqpU1TDnwJYKPVxoq2OIiA/F3JFOFCQghJWCg==")
	assert.Nil(err)
	assert.Nil(stdos.WriteFile(archivePath, archiveBytes, 0o644))

	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      archivePath,
		DestinationInto: "dyd/assets/vendor",
		Unpack:          true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("packed", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "pkg", "contents", "value.txt")))
}

func TestRootRequirementFileBuildStem_UnpackTarXzIntoUsesArchiveName(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "pkg.tar.xz")
	archiveBytes, err := base64.StdEncoding.DecodeString("/Td6WFoAAATm1rRGBMCDAYBQIQEWAAAAAAAAAG1pO4vgJ/8Ae10AMZvKGdrtpRW74LwXrgbt6UOIknkGA/ZH5G5z2wfbhgjFs2Dx1jbZ+G5mUlj6UJ27RMCcS3CCXJTlu8a5lRCCYbK7M0QRklbxI1MhPoVFt2mXir+q2H28l8fQ963CA9vzxN8bdHc/LT6nCh3bcR7i5ymeDyNiQ4a9pCoAAAA6W2I7fa/EuQABnwGAUAAAcMHgk7HEZ/sCAAAAAARZWg==")
	assert.Nil(err)
	assert.Nil(stdos.WriteFile(archivePath, archiveBytes, 0o644))

	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      archivePath,
		DestinationInto: "dyd/assets/vendor",
		Unpack:          true,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("packed", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "pkg", "contents", "value.txt")))
}

func TestRootRequirementFileBuildStem_RejectsUnsafeArchiveSymlink(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "bad.tar")
	archiveFile, err := stdos.Create(archivePath)
	assert.Nil(err)
	tarWriter := tar.NewWriter(archiveFile)
	assert.Nil(tarWriter.WriteHeader(&tar.Header{
		Name:     "escape-link",
		Typeflag: tar.TypeSymlink,
		Linkname: "../../outside",
		Mode:     0o777,
	}))
	assert.Nil(tarWriter.Close())
	assert.Nil(archiveFile.Close())

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:     &SafeGardenReference{BasePath: gardenPath},
		SourcePath: archivePath,
		Unpack:     true,
	})
	assert.NotNil(err)
	assert.Nil(stem)
}

func TestRootRequirementFileBuildStem_RejectsUnsafeZipPath(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	archivePath := filepath.Join(t.TempDir(), "bad.zip")
	archiveFile, err := stdos.Create(archivePath)
	assert.Nil(err)
	zipWriter := zip.NewWriter(archiveFile)
	entry, err := zipWriter.Create("../escape.txt")
	assert.Nil(err)
	_, err = entry.Write([]byte("escape"))
	assert.Nil(err)
	assert.Nil(zipWriter.Close())
	assert.Nil(archiveFile.Close())

	err, stem := RootRequirementFileBuildStem(task.SERIAL_CONTEXT, RootRequirementFileBuildStemRequest{
		Garden:     &SafeGardenReference{BasePath: gardenPath},
		SourcePath: archivePath,
		Unpack:     true,
	})
	assert.NotNil(err)
	assert.Nil(stem)
}
