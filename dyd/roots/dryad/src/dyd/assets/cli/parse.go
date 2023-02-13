// Copyright (c) 2017. Oleg Sklyar & teris.io. All rights reserved.
// See the LICENSE file in the project root for licensing information.

package cli

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
)

const (
	helpKey  = "help"
	helpChar = 'h'
	trueStr  = "true"
)

// Parse parses the original application arguments into the command invocation path (application ->
// first level command -> second level command etc.), a list of validated positional arguments matching
// the command being invoked (the last one in the invocation path) and a map of validated options
// matching one of the invocation path elements, from the top application down to the command being invoked.
// An error is returned if a command is not found or arguments or options are invalid. In case of an error,
// the invocation path is normally also computed and returned (the content of arguments and options is not
// guaranteed). See `App.parse`
func Parse(a App, appargs []string) (invocation []string, args []string, opts map[string]interface{}, err error) {
	_, appname := path.Split(appargs[0])

	invocation, argsAndOpts, expArgs, accptOpts := evalCommand(a, appargs[1:])
	invocation = append([]string{appname}, invocation...)

	if args, opts, err = splitArgsAndOpts(argsAndOpts, accptOpts); err == nil {
		if _, ok := opts["help"]; !ok {
			if err = assertArgs(expArgs, args); err == nil {
				err = assertOpts(accptOpts, opts)
			}
		}
	}
	return invocation, args, opts, err
}

func Unparse(invocation []string, args []string, opts map[string]interface{}) []string {
	var res = make([]string, 0)

	res = append(res, invocation...)
	res = append(res, UnparseOpts(opts)...)
	res = append(res, args...)

	return res
}

func UnparseOpts(opts map[string]interface{}) []string {
	var args = make([]string, 0)

	for key, val := range opts {
		switch val := val.(type) {
		case string:
			args = append(args, fmt.Sprintf("--%s=%s", key, val))
		case bool:
			args = append(args, fmt.Sprintf("--%s=%t", key, val))
		case int64:
			args = append(args, fmt.Sprintf("--%s=%d", key, val))
		case float64:
			args = append(args, fmt.Sprintf("--%s=%f", key, val))
		case []string:
			for _, v := range val {
				args = append(args, fmt.Sprintf("--%s=%s", key, v))
			}
		case []bool:
			for _, v := range val {
				args = append(args, fmt.Sprintf("--%s=%t", key, v))
			}
		case []int64:
			for _, v := range val {
				args = append(args, fmt.Sprintf("--%s=%d", key, v))
			}
		case []float64:
			for _, v := range val {
				args = append(args, fmt.Sprintf("--%s=%f", key, v))
			}
		}
	}

	return args
}

func evalCommand(a App, appargs []string) (invocation []string, argsAndOpts []string, expArgs []Arg, accptOpts []Option) {
	invocation = []string{}
	argsAndOpts = appargs
	expArgs = a.Args()
	accptOpts = a.Options()

	cmds2check := a.Commands()
	for i, arg := range appargs {
		matched := false
		for _, cmd := range cmds2check {
			if cmd.Key() == arg || cmd.Shortcut() == arg {
				invocation = append(invocation, cmd.Key())
				argsAndOpts = appargs[i+1:]
				expArgs = cmd.Args()
				accptOpts = append(accptOpts, cmd.Options()...)

				cmds2check = cmd.Commands()
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}
	return invocation, argsAndOpts, expArgs, accptOpts
}

func setOpt(opts map[string]interface{}, opt Option, raw string) (map[string]interface{}, error) {
	switch opt.Type() {
	case TypeInt:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, err
		}

		opts[opt.Key()] = value
		return opts, nil
	case TypeMultiInt:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, err
		}

		var buffer []int64
		if opts[opt.Key()] != nil {
			buffer = opts[opt.Key()].([]int64)
		} else {
			buffer = make([]int64, 0)
		}
		buffer = append(buffer, value)
		opts[opt.Key()] = buffer

		return opts, nil
	case TypeNumber:
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}

		opts[opt.Key()] = value
		return opts, nil
	case TypeMultiNumber:
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}

		var buffer []float64
		if opts[opt.Key()] != nil {
			buffer = opts[opt.Key()].([]float64)
		} else {
			buffer = make([]float64, 0)
		}
		buffer = append(buffer, value)
		opts[opt.Key()] = buffer

		return opts, nil
	case TypeBool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, err
		}

		opts[opt.Key()] = value
		return opts, nil
	case TypeMultiBool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, err
		}

		var buffer []bool
		if opts[opt.Key()] != nil {
			buffer = opts[opt.Key()].([]bool)
		} else {
			buffer = make([]bool, 0)
		}
		buffer = append(buffer, value)
		opts[opt.Key()] = buffer

		return opts, nil
	case TypeString:
		value := raw

		opts[opt.Key()] = value
		return opts, nil
	case TypeMultiString:
		value := raw

		var buffer []string
		if opts[opt.Key()] != nil {
			buffer = opts[opt.Key()].([]string)
		} else {
			buffer = make([]string, 0)
		}
		buffer = append(buffer, value)
		opts[opt.Key()] = buffer

		return opts, nil
	default:
		return nil, fmt.Errorf("option --%s unhandled by parser", opt.Key())
	}
}

var PARSE_OPT, _ = regexp.Compile(`^(?:--?([^=]+?))(?:=(.+))?$`)

func splitArgsAndOpts(appargs []string, accptOpts []Option) (args []string, opts map[string]interface{}, err error) {
	opts = make(map[string]interface{})

	passthrough := false
	for _, arg := range appargs {
		if arg == "--" {
			passthrough = true
			continue
		}

		if !passthrough {
			matches := PARSE_OPT.FindStringSubmatch(arg)
			matched := matches != nil
			if matched {
				var key string
				var value string
				switch len(matches) {
				case 2:
					key = matches[1]
				case 3:
					key = matches[1]
					value = matches[2]
				}

				if key == helpKey {
					return nil, map[string]interface{}{helpKey: trueStr}, nil
				}

				var opt Option
				for _, accptOpt := range accptOpts {
					if accptOpt.Key() == key {
						opt = accptOpt
					}
				}

				if opt == nil {
					return args, opts, fmt.Errorf("unknown option --%s", key)
				}

				if value == "" {
					if opt.Type() == TypeBool || opt.Type() == TypeMultiBool {
						value = trueStr
					} else {
						return args, opts, fmt.Errorf("missing value for option --%s", key)
					}
				}

				opts, err = setOpt(opts, opt, value)
				if err != nil {
					return args, opts, err
				}
			} else {
				args = append(args, arg)
			}
		} else {
			args = append(args, arg)
		}

		// if !passthrough && strings.HasPrefix(arg, "--") {
		// 	arg = arg[2:]
		// 	if arg == helpKey {
		// 		return nil, map[string]interface{}{helpKey: trueStr}, nil
		// 	}
		// 	parts := strings.Split(arg, "=")
		// 	key := parts[0]
		// 	matched := false
		// 	for _, accptOpt := range accptOpts {
		// 		if accptOpt.Key() == key {
		// 			if accptOpt.Type() == TypeBool {
		// 				if len(parts) == 1 {
		// 					opts, err = setOpt(opts, accptOpt, trueStr)
		// 					if err != nil {
		// 						return args, opts, err
		// 					}
		// 				} else {
		// 					return args, opts, fmt.Errorf("boolean options have true assigned implicitly, found value for --%s", key)
		// 				}
		// 			} else if len(parts) >= 2 {
		// 				opts, err = setOpt(opts, accptOpt, strings.Join(parts[1:], "=")) // permit = in values
		// 				if err != nil {
		// 					return args, opts, err
		// 				}
		// 			} else {
		// 				return args, opts, fmt.Errorf("missing value for option --%s", key)
		// 			}
		// 			matched = true
		// 			break
		// 		}
		// 	}
		// 	if !matched {
		// 		return args, opts, fmt.Errorf("unknown option --%s", key)
		// 	}
		// 	continue
		// }

		// if !passthrough && strings.HasPrefix(arg, "-") {
		// 	arg = arg[1:]

		// 	for i, char := range arg {
		// 		if char == helpChar {
		// 			return nil, map[string]interface{}{helpKey: trueStr}, nil
		// 		}
		// 		matched := false
		// 		for _, accptOpt := range accptOpts {
		// 			if accptOpt.CharKey() == char {
		// 				if accptOpt.Type() == TypeBool {
		// 					opts, err = setOpt(opts, accptOpt, trueStr)
		// 					if err != nil {
		// 						return args, opts, err
		// 					}
		// 				} else if i == len(arg)-1 {
		// 					danglingOpt = accptOpt.Key()
		// 				} else {
		// 					return args, opts, fmt.Errorf("non-boolean flag -%v in non-terminal position", string(char))
		// 				}
		// 				matched = true
		// 				break
		// 			}
		// 		}
		// 		if !matched {
		// 			return args, opts, fmt.Errorf("unknown flag -%v", string(char))
		// 		}
		// 	}
		// 	continue
		// }

	}

	return args, opts, nil
}

func assertArgs(expected []Arg, actual []string) error {
	if len(expected) == 0 || !expected[len(expected)-1].Optional() {
		if len(expected) > len(actual) {
			return fmt.Errorf("missing required argument %v", expected[len(actual)].Key())
		} else if len(expected) < len(actual) {
			return fmt.Errorf("unknown arguments %v", actual[len(expected):])
		}
	}
	for i, e := range expected {
		if len(actual) < i+1 {
			if !e.Optional() {
				return fmt.Errorf("missing required argument %s", e.Key())
			}
			break
		}
		arg := actual[i]
		switch e.Type() {
		case TypeBool:
			if _, err := strconv.ParseBool(arg); err != nil {
				return fmt.Errorf("argument %s must be a boolean value, found %v", e.Key(), arg)
			}
		case TypeInt:
			if _, err := strconv.ParseInt(arg, 10, 64); err != nil {
				return fmt.Errorf("argument %s must be an integer value, found %v", e.Key(), arg)
			}
		case TypeNumber:
			if _, err := strconv.ParseFloat(arg, 64); err != nil {
				return fmt.Errorf("argument %s must be a number, found %v", e.Key(), arg)
			}
		}
	}
	return nil
}

func assertOpt(option Option, value interface{}) error {
	switch option.Type() {
	case TypeInt:
		if _, isType := value.(int64); !isType {
			return fmt.Errorf("option --%s must be an integer", option.Key())
		}
	case TypeMultiInt:
		if _, isType := value.([]int64); !isType {
			return fmt.Errorf("option --%s must be an integer array", option.Key())
		}
	case TypeNumber:
		if _, isType := value.(float64); !isType {
			return fmt.Errorf("option --%s must must be a number", option.Key())
		}
	case TypeMultiNumber:
		if _, isType := value.([]float64); !isType {
			return fmt.Errorf("option --%s must must be a number array", option.Key())
		}
	case TypeBool:
		if _, isType := value.(bool); !isType {
			return fmt.Errorf("option --%s must must be a boolean", option.Key())
		}
	case TypeMultiBool:
		if _, isType := value.([]bool); !isType {
			return fmt.Errorf("option --%s must must be a boolean array", option.Key())
		}
	case TypeString:
		if _, isType := value.(string); !isType {
			return fmt.Errorf("option --%s must must be a string", option.Key())
		}
	case TypeMultiString:
		if _, isType := value.([]string); !isType {
			return fmt.Errorf("option --%s must must be a string array", option.Key())
		}
	default:
		return fmt.Errorf("option --%s unhandled by validator", option.Key())
	}
	return nil
}

func assertOpts(permitted []Option, actual map[string]interface{}) error {
	for key, value := range actual {
		for _, p := range permitted {
			if p.Key() == key {
				if err := assertOpt(p, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
