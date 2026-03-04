package exec

import (
	"context"
	"dryad/diagnostics"
	stdexec "os/exec"
)

type Cmd struct {
	*stdexec.Cmd
}

func commandKey(cmd *Cmd) string {
	if cmd.Path != "" {
		return cmd.Path
	}
	if len(cmd.Args) > 0 {
		return cmd.Args[0]
	}
	return ""
}

var run = diagnostics.BindA1R0(
	"exec.run",
	commandKey,
	func(cmd *Cmd) error {
		return cmd.Cmd.Run()
	},
)

var start = diagnostics.BindA1R0(
	"exec.start",
	commandKey,
	func(cmd *Cmd) error {
		return cmd.Cmd.Start()
	},
)

var wait = diagnostics.BindA1R0(
	"exec.wait",
	commandKey,
	func(cmd *Cmd) error {
		return cmd.Cmd.Wait()
	},
)

func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd: stdexec.Command(name, arg...),
	}
}

func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd: stdexec.CommandContext(ctx, name, arg...),
	}
}

func (cmd *Cmd) Run() error {
	return run(cmd)
}

func (cmd *Cmd) Start() error {
	return start(cmd)
}

func (cmd *Cmd) Wait() error {
	return wait(cmd)
}
