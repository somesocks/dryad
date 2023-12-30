package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var ScopedCommand = func(
	command clib.Command,
) clib.Command {
	action := command.Action()

	wrapper := func(req clib.ActionRequest) int {
		invocation := req.Invocation
		options := req.Opts

		var scope string
		if options["scope"] != nil {
			scope = options["scope"].(string)
		} else {
			var err error
			scope, err = dryad.ScopeGetDefault(scope)
			zlog.Info().Msg("loading default scope: " + scope)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while loading default scope")
				return 1
			}
		}

		// if the scope is unset, bypass expansion and run the action directly
		if scope == "" || scope == "none" {
			return action(req)
		} else {
			zlog.Info().Msg("using scope: " + scope)
		}

		settingName := strings.Join(invocation[1:], "-")

		path, err := os.Getwd()
		if err != nil {
			zlog.Fatal().Err(err).Msg("error while finding working directory")
			return 1
		}

		setting, err := dryad.ScopeSettingGet(path, scope, settingName)
		if err != nil {
			zlog.Fatal().Err(err).Msg("error while loading setting")
			return 1
		}

		settings, err := dryad.ScopeSettingParseShell(setting)
		if err != nil {
			zlog.Fatal().Err(err).Msg("error while parsing setting")
			return 1
		}

		// if the scope setting is unset,
		// there`s no need to modify the request
		if len(settings) == 0 {
			return action(req)
		}

		argsRewrite := make([]string, 0)
		index := 0

		// copy all of the arguments before the scope arg
		for index < len(os.Args) {
			var element = os.Args[index]
			index++ // do this here so it increments before the break
			if strings.HasPrefix(element, "--scope=") {
				break
			} else {
				argsRewrite = append(argsRewrite, element)
			}
		}

		// insert a null scope arg, to stop a loop
		argsRewrite = append(argsRewrite, "--scope=none")

		// insert the new args from the settings in place of the scope
		argsRewrite = append(argsRewrite, settings...)

		// copy all of the arguments after the scope arg
		for index < len(os.Args) {
			var element = os.Args[index]
			argsRewrite = append(argsRewrite, element)
			index++
		}
		zlog.Debug().Msg("rewriting args to: " + strings.Join(argsRewrite, " "))
		return req.App.Run(argsRewrite, os.Stdout)
	}

	return command.
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithAction(wrapper)
}
