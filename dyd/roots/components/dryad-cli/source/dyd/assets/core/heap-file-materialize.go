package core

import (
	"dryad/internal/os"
	"dryad/internal/unix"

	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"syscall"
)

const (
	heapRolloverMaxRetries = 8
)

func heapMaterializeFile(sourcePath string, destPath string) error {
	err := os.Link(sourcePath, destPath)
	if err == nil {
		return nil
	}
	if !errors.Is(err, syscall.EMLINK) {
		return err
	}
	if err := heapMaterializeFilePreserveExistingLink(destPath, err); err != nil {
		return err
	}

	return heapMaterializeFileWithRollover(sourcePath, destPath)
}

func heapMaterializeFileWithRollover(sourcePath string, destPath string) (err error) {
	lockFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer lockFile.Close()

	err = unix.Flock(int(lockFile.Fd()), unix.LOCK_EX)
	if err != nil {
		return err
	}
	defer unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)

	err = os.Link(sourcePath, destPath)
	if err == nil {
		return nil
	}
	if !errors.Is(err, syscall.EMLINK) {
		return err
	}
	if err := heapMaterializeFilePreserveExistingLink(destPath, err); err != nil {
		return err
	}

	for attempt := 0; attempt < heapRolloverMaxRetries; attempt++ {
		tempPath, err := heapCreateMaterializationTemp(sourcePath)
		if err != nil {
			return err
		}

		err = os.Link(tempPath, destPath)
		if err == nil {
			err = os.Rename(tempPath, sourcePath)
			if err != nil {
				_ = os.Remove(tempPath)
				return err
			}
			return nil
		}

		_ = os.Remove(tempPath)
		if !errors.Is(err, syscall.EMLINK) {
			return err
		}
		if err := heapMaterializeFilePreserveExistingLink(destPath, err); err != nil {
			return err
		}
	}

	return fmt.Errorf("heap file rollover exhausted retries for %q", sourcePath)
}

func heapMaterializeFilePreserveExistingLink(destPath string, linkErr error) error {
	if _, statErr := os.Lstat(destPath); statErr == nil {
		return linkErr
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return statErr
	}

	return nil
}

func heapCreateMaterializationTemp(sourcePath string) (tempPath string, err error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", err
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	tempFile, err := os.CreateTemp(
		filepath.Dir(sourcePath),
		".tmp-"+filepath.Base(sourcePath)+"-*",
	)
	if err != nil {
		return "", err
	}

	tempPath = tempFile.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()

	_, err = io.Copy(tempFile, sourceFile)
	if err != nil {
		tempFile.Close()
		return "", err
	}

	err = tempFile.Chmod(sourceInfo.Mode().Perm())
	if err != nil {
		tempFile.Close()
		return "", err
	}

	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	return tempPath, nil
}
