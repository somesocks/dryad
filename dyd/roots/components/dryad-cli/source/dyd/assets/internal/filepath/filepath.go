package filepath

import (
	"dryad/diagnostics"
	"io/fs"
	stdfilepath "path/filepath"
)

var glob = diagnostics.BindA1R1(
	"filepath.glob",
	func(pattern string) string {
		return pattern
	},
	func(pattern string) (error, []string) {
		matches, err := stdfilepath.Glob(pattern)
		return err, matches
	},
)

func Glob(pattern string) ([]string, error) {
	err, matches := glob(pattern)
	return matches, err
}

var WalkDir = diagnostics.BindA2R0(
	"filepath.walk_dir",
	func(root string, _ fs.WalkDirFunc) string {
		return root
	},
	stdfilepath.WalkDir,
)
