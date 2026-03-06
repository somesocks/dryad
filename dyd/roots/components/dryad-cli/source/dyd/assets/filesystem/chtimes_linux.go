//go:build linux || android

// this package implements chtimes without following symlinks

package fs2

import (
	"dryad/internal/unix"
	"time"
)

func Chtimes(path string, at time.Time, mt time.Time) error {
	ts := []unix.Timespec{
		unix.NsecToTimespec(at.UnixNano()),
		unix.NsecToTimespec(mt.UnixNano()),
	}
	return unix.UtimesNanoAt(unix.AT_FDCWD, path, ts, unix.AT_SYMLINK_NOFOLLOW)
}
