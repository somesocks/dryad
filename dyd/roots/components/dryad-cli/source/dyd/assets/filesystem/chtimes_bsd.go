//go:build darwin || freebsd || netbsd || openbsd

package fs2

import (
	"time"
	"golang.org/x/sys/unix"
)

func Chtimes(path string, at time.Time, mt time.Time) error {
	ts := []unix.Timespec{
		unix.NsecToTimespec(at.UnixNano()),
		unix.NsecToTimespec(mt.UnixNano()),
	}
	if err := unix.UtimesNanoAt(unix.AT_FDCWD, path, ts, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		if err != unix.ENOSYS {
			return err
		}

		// Older macOS: fall back to Lutimes
		tv := []unix.Timeval{
			unix.NsecToTimeval(at.UnixNano()),
			unix.NsecToTimeval(mt.UnixNano()),
		}
		return unix.Lutimes(path, tv)
	}
	return nil
}
