package unix

import (
	"dryad/diagnostics"
	stdunix "golang.org/x/sys/unix"
)

type Timespec = stdunix.Timespec
type Timeval = stdunix.Timeval

const (
	LOCK_EX            = stdunix.LOCK_EX
	LOCK_UN            = stdunix.LOCK_UN
	AT_FDCWD           = stdunix.AT_FDCWD
	AT_SYMLINK_NOFOLLOW = stdunix.AT_SYMLINK_NOFOLLOW
	ENOSYS             = stdunix.ENOSYS
)

func NsecToTimespec(nsec int64) Timespec {
	return stdunix.NsecToTimespec(nsec)
}

func NsecToTimeval(nsec int64) Timeval {
	return stdunix.NsecToTimeval(nsec)
}

var Flock = diagnostics.BindA2R0(
	"unix.flock",
	nil,
	stdunix.Flock,
)

type utimesNanoAtRequest struct {
	DirFD int
	Path  string
	TS    []Timespec
	Flags int
}

var utimesNanoAt = diagnostics.BindA1R0(
	"unix.utimes_nano_at",
	func(req utimesNanoAtRequest) string {
		return req.Path
	},
	func(req utimesNanoAtRequest) error {
		return stdunix.UtimesNanoAt(req.DirFD, req.Path, req.TS, req.Flags)
	},
)

func UtimesNanoAt(dirfd int, path string, ts []Timespec, flags int) error {
	return utimesNanoAt(utimesNanoAtRequest{
		DirFD: dirfd,
		Path:  path,
		TS:    ts,
		Flags: flags,
	})
}

var Lutimes = diagnostics.BindA2R0(
	"unix.lutimes",
	func(path string, _ []Timeval) string {
		return path
	},
	stdunix.Lutimes,
)
