package os

import (
	"dryad/diagnostics"
	stdos "os"
	"time"
)

var Chtimes = diagnostics.BindA3R0(
	"os.chtimes",
	func(name string, _ time.Time, _ time.Time) string {
		return name
	},
	stdos.Chtimes,
)
