// Copyright (c) 2017. Oleg Sklyar & teris.io. All rights reserved.
// See the LICENSE file in the project root for licensing information.

// Package cli provides a simple, fast and complete API for building command line applications in Go.
// In contrast to other libraries additional emphasis is put on the definition and validation of
// positional arguments and consistent usage outputs combining options from all command levels into
// one block.
package cli_builder

import (
	"fmt"
	"io"
	"strings"
)

type ActionRequest struct {
	App        App
	Invocation []string
	Args       []string
	Opts       map[string]any
}

type AutoCompleteRequest struct {
	App    App
	Tokens []string
}

// Action defines a function type to be executed for an application or a
// command. It takes a slice of validated positional arguments and a map
// of validated options (with all value types encoded as strings) and
// returns a Unix exit code (success: 0).
type Action func(request ActionRequest) int

type AutoComplete func(request AutoCompleteRequest) []string

// App defines a CLI application parameterizable with sub-commands, arguments and options.
type App interface {
	// Description returns the application description to be output in the usage.
	Description() string
	// Args returns required and optional positional arguments for the top-level application.
	Args() []Arg
	// Options permitted for the top-level application and all sub-commands.
	Options() []Option
	// Commands returns the set of first-level sub-commands for the application.
	Commands() []Command
	// Action returns the application action when no sub-command is specified.
	Action() Action

	// WithArg adds a positional argument to the application. Specifying last application/command
	// argument as optional permits unlimited number of further positional arguments (at least one
	// optional argument needs to be specified in the definition for this case).
	WithArg(arg Arg) App
	// WithOption adds a permitted option to the application and all sub-commands.
	WithOption(opt Option) App
	// WithCommand adds a first-level sub-command to the application.
	WithCommand(cmd Command) App
	// WithAction sets the action function to execute after successful parsing of commands, arguments
	// and options to the top-level application.
	WithAction(action Action) App

	// Parse parses the original application arguments into the command invocation path (application ->
	// first level command -> second level command etc.), a list of validated positional arguments matching
	// the command being invoked (the last one in the invocation path) and a map of validated options
	// matching one of the invocation path elements, from the top application down to the command being invoked.
	// An error is returned if a command is not found or arguments or options are invalid. In case of an error,
	// the invocation path is normally also computed and returned (the content of arguments and options is not
	// guaranteed).
	Parse(appargs []string) (invocation []string, args []string, opts map[string]interface{}, err error)
	// Run parses the argument list and runs the command specified with the corresponding options and arguments.
	Run(appargs []string, w io.Writer) int
	// Usage prints out the full usage help.
	Usage(invocation []string, w io.Writer) error

	AutoComplete(tokens []string) (error, []string)
}

// New creates a new CLI App.
func New(descr string) App {
	return &app{descr: descr}
}

type app struct {
	descr        string
	args         []Arg
	opts         []Option
	cmds         []Command
	action       Action
	autoComplete AutoComplete
}

func (a *app) Description() string {
	return a.descr
}

func (a *app) Args() []Arg {
	return a.args
}

func (a *app) Options() []Option {
	return a.opts
}

func (a *app) Commands() []Command {
	return a.cmds
}

func (a *app) Action() Action {
	return a.action
}

func (a *app) AutoComplete(tokens []string) (error, []string) {
	return AppAutoComplete(a, tokens)
}

func (a *app) WithArg(arg Arg) App {
	a.args = append(a.args, arg)
	return a
}

func (a *app) WithOption(opt Option) App {
	a.opts = append(a.opts, opt)
	return a
}

func (a *app) WithCommand(cmd Command) App {
	a.cmds = append(a.cmds, cmd)
	return a
}
func (a *app) WithAction(action Action) App {
	a.action = action
	return a
}

func (a *app) Parse(appargs []string) (invocation []string, args []string, opts map[string]interface{}, err error) {
	return Parse(a, appargs)
}

func (a *app) Run(appargs []string, w io.Writer) int {
	invocation, args, opts, err := a.Parse(appargs)
	_, help := opts["help"]
	code := 1
	if help && (err == nil || strings.HasPrefix(err.Error(), "missing required argument ") ) {
		a.Usage(invocation, w)
		code = 0
	} else if err != nil {
		fmt.Fprintf(w, "fatal: %v\n", err)
		fmt.Fprintf(w, "usage: %v\n", shortUsage(a, invocation))
	} else {
		action := a.Action()
		if len(invocation) > 1 {
			cmds := a.Commands()
			for i, key := range invocation[1:] {
				matched := false
				for _, cmd := range cmds {
					if cmd.Key() == key {
						cmds = cmd.Commands()
						matched = true
						if i == len(invocation)-2 {
							action = cmd.Action()
						}
						break
					}
				}
				// should never happen if invocation originates from the parser
				if !matched {
					fmt.Fprintf(w, "fatal: invalid invocation path %v\n", invocation)
					fmt.Fprintf(w, "usage: %v\n", shortUsage(a, invocation[:1]))
					action = nil
					break
				}
			}
		}
		if action != nil {
			code = action(ActionRequest{
				App:        a,
				Invocation: invocation,
				Args:       args,
				Opts:       opts,
			})
		} else {
			a.Usage(invocation, w)
			code = 1
		}
	}
	return code
}

func (a *app) Usage(invocation []string, w io.Writer) error {
	return Usage(a, invocation, w)
}
