package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

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

			conn, err := net.Dial("unix", socketPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to connect to root develop socket")
				return 1
			}
			defer conn.Close()

			_, err = conn.Write([]byte("status\n"))
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to send status request")
				return 1
			}

			reader := bufio.NewReader(conn)
			line, err := reader.ReadString('\n')
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed to read status response")
				return 1
			}

			status := strings.TrimSpace(line)
			if status != "ok" {
				zlog.Fatal().Err(errors.New(status)).Msg("status request failed")
				return 1
			}

			for {
				line, err = reader.ReadString('\n')
				if err != nil {
					zlog.Fatal().Err(err).Msg("failed to read status response")
					return 1
				}
				text := strings.TrimSpace(line)
				if text == "end" {
					break
				}
				fmt.Println(text)
			}

			return 0
		})

	command = LoggingCommand(command)

	return command
}()
