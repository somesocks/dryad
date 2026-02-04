package cli

import (
	clib "dryad/cli-builder"
	"errors"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopSaveCommand = func() clib.Command {
	command := clib.NewCommand("save", "save changes from an active root development environment").
		WithAction(func(req clib.ActionRequest) int {
			socketPath := os.Getenv("DYD_DEV_SOCKET")
			if socketPath == "" {
				zlog.Fatal().Err(errors.New("DYD_DEV_SOCKET not set")).Msg("not running inside a root development environment")
				return 1
			}

			res, err := rootDevelopIPC_send(socketPath, "save")
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to send save request")
				return 1
			}
			if res.Status != "ok" {
				msg := res.Message
				if msg == "" {
					msg = "save request failed"
				}
				zlog.Fatal().Err(errors.New(msg)).Msg("save request failed")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)

	return command
}()
