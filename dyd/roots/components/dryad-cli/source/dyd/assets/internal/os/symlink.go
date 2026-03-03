package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Symlink = diagnostics.BindA2R0(
	"os.symlink",
	func(target string, linkPath string) string {
		return linkPath
	},
	stdos.Symlink,
)
