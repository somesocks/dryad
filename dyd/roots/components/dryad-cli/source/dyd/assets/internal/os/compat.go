package os

import stdos "os"

type File = stdos.File
type FileInfo = stdos.FileInfo
type FileMode = stdos.FileMode
type DirEntry = stdos.DirEntry

const (
	ModePerm          = stdos.ModePerm
	ModeSymlink       = stdos.ModeSymlink
	PathSeparator     = stdos.PathSeparator
	PathListSeparator = stdos.PathListSeparator
	O_RDONLY          = stdos.O_RDONLY
	O_WRONLY          = stdos.O_WRONLY
	O_RDWR            = stdos.O_RDWR
	O_APPEND          = stdos.O_APPEND
	O_CREATE          = stdos.O_CREATE
	O_EXCL            = stdos.O_EXCL
	O_SYNC            = stdos.O_SYNC
	O_TRUNC           = stdos.O_TRUNC
)

var (
	Args           = stdos.Args
	Stdin          = stdos.Stdin
	Stdout         = stdos.Stdout
	Stderr         = stdos.Stderr
	ErrNotExist    = stdos.ErrNotExist
	ErrProcessDone = stdos.ErrProcessDone
	Interrupt      = stdos.Interrupt
)

func Environ() []string {
	return stdos.Environ()
}

func Exit(code int) {
	stdos.Exit(code)
}

func Getenv(key string) string {
	return stdos.Getenv(key)
}

func IsNotExist(err error) bool {
	return stdos.IsNotExist(err)
}

func UserHomeDir() (string, error) {
	return stdos.UserHomeDir()
}
