package cli

import (
	clib "dryad/cli-builder"
	"errors"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopSnapshotCommand = func() clib.Command {
	command := clib.NewCommand("snapshot", "snapshot the current development workspace state").
		WithAction(func(req clib.ActionRequest) int {
			socketPath := os.Getenv("DYD_DEV_SOCKET")
			if socketPath == "" {
				zlog.Fatal().Err(errors.New("DYD_DEV_SOCKET not set")).Msg("not running inside a root development environment")
				return 1
			}

			res, err := rootDevelopIPC_send(socketPath, "snapshot")
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to send snapshot request")
				return 1
			}
			if res.Status != "ok" {
				msg := res.Message
				if msg == "" {
					msg = "snapshot request failed"
				}
				zlog.Fatal().Err(errors.New(msg)).Msg("snapshot request failed")
				return 1
			}

			if res.Message != "" {
				fmt.Printf("snapshot saved %s\n", res.Message)
			} else {
				fmt.Println("snapshot saved")
			}

			return 0
		})

	command = LoggingCommand(command)

	return command
}()
