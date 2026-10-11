package core

import "fmt"

type InputMode int

const (
	InputAuto InputMode = iota
	InputNone
	InputInherit
)

type InputConfig struct {
	Mode InputMode
}

// Resolve chooses stdin ownership for the entire invocation, before any child
// starts. Parallelism refers to the configured limit, not current activity.
func (config InputConfig) Resolve(executions, parallel int, fromStdin bool) (InputConfig, error) {
	if executions < 0 || parallel < 1 {
		return InputConfig{}, fmt.Errorf("invalid execution count or parallelism for stdin")
	}
	concurrent := executions > 1 && parallel > 1
	switch config.Mode {
	case InputAuto:
		if fromStdin || concurrent {
			return InputConfig{Mode: InputNone}, nil
		}
		return InputConfig{Mode: InputInherit}, nil
	case InputNone:
		return config, nil
	case InputInherit:
		if fromStdin {
			return InputConfig{}, fmt.Errorf("--join-stdin=true cannot be used with --from-stdin")
		}
		if concurrent {
			return InputConfig{}, fmt.Errorf("--join-stdin=true requires --parallel=1 when multiple executions are selected")
		}
		return config, nil
	default:
		return InputConfig{}, fmt.Errorf("invalid input mode: %d", config.Mode)
	}
}

type OutputMode int

const (
	OutputInherit OutputMode = iota
	OutputNone
	OutputLog
)

type OutputConfig struct {
	Mode OutputMode
	Path string
	Name string
}

// OutputConfigFromOptions preserves the join/log precedence used by existing
// callers: joining wins when enabled; otherwise a nonempty log path enables logging.
func OutputConfigFromOptions(join bool, path, name string) OutputConfig {
	if join {
		return OutputConfig{Mode: OutputInherit}
	}
	if path != "" {
		return OutputConfig{Mode: OutputLog, Path: path, Name: name}
	}
	return OutputConfig{Mode: OutputNone}
}

func (config OutputConfig) validate() error {
	switch config.Mode {
	case OutputInherit, OutputNone:
		if config.Path != "" || config.Name != "" {
			return fmt.Errorf("output path and name require log mode")
		}
	case OutputLog:
		if config.Path == "" {
			return fmt.Errorf("output log mode requires a directory")
		}
	default:
		return fmt.Errorf("invalid output mode: %d", config.Mode)
	}
	return nil
}
