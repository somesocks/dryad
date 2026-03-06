package core

import (
	dfilepath "dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/internal/unix"

	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	heapReplicaPathSeparator = ".replica."
	heapReplicaMaxRetries    = 8
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

	return heapMaterializeFileWithReplicas(sourcePath, destPath)
}

func heapMaterializeFileWithReplicas(sourcePath string, destPath string) (err error) {
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

	for attempt := 0; attempt < heapReplicaMaxRetries; attempt++ {
		replicas, err := dfilepath.Glob(sourcePath + heapReplicaPathSeparator + "*")
		if err != nil {
			return err
		}

		for _, replicaPath := range replicas {
			err = os.Link(replicaPath, destPath)
			if err == nil {
				return nil
			}
			if !errors.Is(err, syscall.EMLINK) {
				return err
			}
			if err := heapMaterializeFilePreserveExistingLink(destPath, err); err != nil {
				return err
			}
		}

		replicaPath, err := heapNextReplicaPath(sourcePath, replicas)
		if err != nil {
			return err
		}

		err = heapCreateReplica(sourcePath, replicaPath)
		if err != nil {
			return err
		}

		err = os.Link(replicaPath, destPath)
		if err == nil {
			return nil
		}
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

func heapNextReplicaPath(sourcePath string, replicas []string) (string, error) {
	maxIndex := 0

	for _, replicaPath := range replicas {
		index, err := heapReplicaPathIndex(sourcePath, replicaPath)
		if err != nil {
			return "", err
		}
		if index > maxIndex {
			maxIndex = index
		}
	}

	return sourcePath + heapReplicaPathSeparator + strconv.Itoa(maxIndex+1), nil
}

func heapReplicaPathIndex(sourcePath string, replicaPath string) (int, error) {
	prefix := sourcePath + heapReplicaPathSeparator
	if !strings.HasPrefix(replicaPath, prefix) {
		return 0, fmt.Errorf("invalid heap replica path %q for %q", replicaPath, sourcePath)
	}

	index, err := strconv.Atoi(strings.TrimPrefix(replicaPath, prefix))
	if err != nil {
		return 0, fmt.Errorf("invalid heap replica index for %q: %w", replicaPath, err)
	}
	if index <= 0 {
		return 0, fmt.Errorf("invalid heap replica index for %q", replicaPath)
	}

	return index, nil
}

func heapCreateReplica(sourcePath string, replicaPath string) (err error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	tempFile, err := os.CreateTemp(
		filepath.Dir(sourcePath),
		".tmp-"+filepath.Base(replicaPath)+"-*",
	)
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	_, err = io.Copy(tempFile, sourceFile)
	if err != nil {
		tempFile.Close()
		return err
	}

	err = tempFile.Chmod(sourceInfo.Mode().Perm())
	if err != nil {
		tempFile.Close()
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	return os.Rename(tempPath, replicaPath)
}
