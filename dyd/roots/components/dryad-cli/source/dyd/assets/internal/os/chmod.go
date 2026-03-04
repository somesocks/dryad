package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Chmod = diagnostics.BindA2R0(
	"os.chmod",
	func(name string, _ stdos.FileMode) string {
		return name
	},
	stdos.Chmod,
)
