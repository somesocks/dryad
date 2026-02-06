package cli

import (
	clib "dryad/cli-builder"
	"dryad/task"
	"errors"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopStatusCommand = func() clib.Command {
	var errDevSocketNotSet = errors.New("DYD_DEV_SOCKET not set")

	type ParsedArgs struct {
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var opts = req.Opts
		var parallel int

		if opts["parallel"] != nil {
			parallel = int(opts["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Parallel: parallel,
		}
	}

	var runStatus = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		socketPath := os.Getenv("DYD_DEV_SOCKET")
		if socketPath == "" {
			return errDevSocketNotSet, nil
		}

		res, err := rootDevelopIPC_send(socketPath, "status")
		if err != nil {
			return err, nil
		}
		if res.Status != "ok" {
			msg := res.Message
			if msg == "" {
				msg = "status request failed"
			}
			return errors.New(msg), nil
		}

		for _, entry := range res.Entries {
			fmt.Println(entry.Code + " " + entry.Path)
		}

		return nil, nil
	}

	runStatus = task.WithContext(
		runStatus,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runStatus,
		),
		func(err error, val any) int {
			if err != nil {
				if errors.Is(err, errDevSocketNotSet) {
					zlog.Fatal().Err(err).Msg("not running inside a root development environment")
					return 1
				}
				zlog.Fatal().Err(err).Msg("status request failed")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("status", "show the status of an active root development environment").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
