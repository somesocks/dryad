package fs2

import (
	// "errors"
	"os"
	// "runtime"

	"golang.org/x/sys/unix"
	// "golang.org/x/sys/windows"

	// "dryad/task"
)

type FileLock struct {
	file *os.File
}

func newFileLock(file *os.File) FileLock {
	return FileLock{
		file: file,
	}
}

// Lock acquires an exclusive lock on the file.
func (fl *FileLock) Lock() error {
	// if runtime.GOOS == "windows" {
	// 	return lockFileWindows(fl.file)
	// }
	return lockFileUnix(fl.file)
}

// Unlock releases the lock on the file.
func (fl *FileLock) Unlock() error {
	// if runtime.GOOS == "windows" {
	// 	return unlockFileWindows(fl.file)
	// }
	return unlockFileUnix(fl.file)
}

// lockFileUnix locks a file on Unix-based systems.
func lockFileUnix(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX)
}

// unlockFileUnix unlocks a file on Unix-based systems.
func unlockFileUnix(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}

// // lockFileWindows locks a file on Windows systems.
// func lockFileWindows(file *os.File) error {
// 	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // unlockFileWindows unlocks a file on Windows systems.
// func unlockFileWindows(file *os.File) error {
// 	err := windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
