package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
)

func inputConfigFromOptions(options map[string]any) dryad.InputConfig {
	if value := options["join-stdin"]; value != nil {
		if value.(bool) {
			return dryad.InputConfig{Mode: dryad.InputInherit}
		}
		return dryad.InputConfig{Mode: dryad.InputNone}
	}
	return dryad.InputConfig{Mode: dryad.InputAuto}
}

func joinStdinOption() clib.Option {
	return clib.NewOption("join-stdin", "provide parent stdin to children. true requires --parallel=1 for multiple executions and cannot be used with --from-stdin. false provides EOF. default: inherit for single/serial execution, EOF otherwise").WithType(clib.OptionTypeBool)
}
