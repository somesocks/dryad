package cli

import (
	clib "dryad/cli-builder"
	"os"

	zerolog "github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
)

var LoggingCommand = func(
	command clib.Command,
) clib.Command {
	action := command.Action()

	wrapper := func(req clib.ActionRequest) int {
		var options = req.Opts
		var loglevel string

		if options["log-level"] != nil {
			loglevel = options["log-level"].(string)
		} else {
			loglevel = "info"
		}


		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

		switch loglevel {
		case "panic":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "fatal":
			zerolog.SetGlobalLevel(zerolog.FatalLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "trace":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		default:
			log.Fatal().Msg("unrecognized log level " + loglevel)
			return 1
		}

		return action(req)
	}

	return command.
		WithOption(clib.NewOption("log-level", "set the logging level").WithType(clib.OptionTypeString)).
		WithAction(wrapper)
}
