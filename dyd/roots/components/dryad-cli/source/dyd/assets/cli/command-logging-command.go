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
		var logFormat string
		var logLevel string

		if options["log-level"] != nil {
			logLevel = options["log-level"].(string)
		} else {
			logLevel = "info"
		}

		if options["log-format"] != nil {
			logFormat = options["log-format"].(string)
		} else {
			logFormat = "console"
		}

		switch logFormat {
		case "console":
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
		case "json":
			break
		default:
			log.Fatal().Msg("unrecognized log format " + logFormat)
			return 1
		}

		switch logLevel {
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
			log.Fatal().Msg("unrecognized log level " + logLevel)
			return 1
		}

		return action(req)
	}

	return command.
		WithOption(clib.NewOption("log-level", "set the logging level. can be one of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', or 'trace'.  defaults to 'info'").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("log-format", "set the logging format. can be one of 'console' or 'json'.  defaults to 'console'").WithType(clib.OptionTypeString)).
		WithAction(wrapper)
}
