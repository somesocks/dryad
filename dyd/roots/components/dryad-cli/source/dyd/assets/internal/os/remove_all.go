package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var RemoveAll = diagnostics.BindA1R0(
	"os.remove_all",
	func(path string) string {
		return path
	},
	stdos.RemoveAll,
)
