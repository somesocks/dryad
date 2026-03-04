package process

import (
	"dryad/diagnostics"
	"dryad/internal/os"
	"fmt"
)

var Signal = diagnostics.BindA2R0(
	"process.signal",
	func(proc *os.Process, signal os.Signal) string {
		if proc == nil {
			return fmt.Sprintf("nil:%v", signal)
		}
		return fmt.Sprintf("%d:%v", proc.Pid, signal)
	},
	func(proc *os.Process, signal os.Signal) error {
		return proc.Signal(signal)
	},
)
