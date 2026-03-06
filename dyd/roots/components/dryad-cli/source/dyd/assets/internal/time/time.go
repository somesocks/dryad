package time

import (
	"dryad/diagnostics"
	stdtime "time"
)

var now = diagnostics.BindA0R1(
	"time.now",
	func() (error, stdtime.Time) {
		return nil, stdtime.Now()
	},
)

func Now() stdtime.Time {
	_, t := now()
	return t
}
