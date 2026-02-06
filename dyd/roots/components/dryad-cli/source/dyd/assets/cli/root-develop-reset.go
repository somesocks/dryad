package cli

import (
	clib "dryad/cli-builder"
	"errors"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopResetCommand = func() clib.Command {
	command := clib.NewCommand("reset", "reset the development workspace to the snapshot state").
		WithAction(func(req clib.ActionRequest) int {
			socketPath := os.Getenv("DYD_DEV_SOCKET")
			if socketPath == "" {
				zlog.Fatal().Err(errors.New("DYD_DEV_SOCKET not set")).Msg("not running inside a root development environment")
				return 1
			}

			res, err := rootDevelopIPC_send(socketPath, "reset")
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to send reset request")
				return 1
			}
			if res.Status != "ok" {
				msg := res.Message
				if msg == "" {
					msg = "reset request failed"
				}
				zlog.Fatal().Err(errors.New(msg)).Msg("reset request failed")
				return 1
			}

			fmt.Println("reset complete")

			return 0
		})

	command = LoggingCommand(command)

	return command
}()
