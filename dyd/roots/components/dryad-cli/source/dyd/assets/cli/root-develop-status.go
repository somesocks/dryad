package cli

import (
	clib "dryad/cli-builder"
	"errors"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopStatusCommand = func() clib.Command {
	command := clib.NewCommand("status", "show the status of an active root development environment").
		WithAction(func(req clib.ActionRequest) int {
			socketPath := os.Getenv("DYD_DEV_SOCKET")
			if socketPath == "" {
				zlog.Fatal().Err(errors.New("DYD_DEV_SOCKET not set")).Msg("not running inside a root development environment")
				return 1
			}

			res, err := rootDevelopIPC_send(socketPath, "status")
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to send status request")
				return 1
			}
			if res.Status != "ok" {
				msg := res.Message
				if msg == "" {
					msg = "status request failed"
				}
				zlog.Fatal().Err(errors.New(msg)).Msg("status request failed")
				return 1
			}

			for _, entry := range res.Entries {
				fmt.Println(entry.Code + " " + entry.Path)
			}

			return 0
		})

	command = LoggingCommand(command)

	return command
}()
