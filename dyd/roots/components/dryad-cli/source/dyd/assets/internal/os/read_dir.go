package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var readDir = diagnostics.BindA1R1(
	"os.read_dir",
	func(name string) string {
		return name
	},
	func(name string) (error, []stdos.DirEntry) {
		entries, err := stdos.ReadDir(name)
		return err, entries
	},
)

var ReadDir = func(name string) ([]stdos.DirEntry, error) {
	err, entries := readDir(name)
	return entries, err
}
