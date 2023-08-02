package cli_builder

type ArgAutoComplete func(token string) []string

// Arg defines a positional argument. Arguments are validated for their
// count and their type. If the last defined argument is optional, then
// an unlimited number of arguments can be passed into the call, otherwise
// an exact count of positional arguments is expected. Obviously, optional
// arguments can be omitted. No validation is done for the invalid case of
// specifying an optional positional argument before a required one.
type Arg interface {

	// Key defines how the argument will be shown in the usage string.
	Key() string

	// Description returns the description of the argument usage
	Description() string

	// Type defines argument type. Default is string, which is not validated,
	// other types are validated by simple string parsing into boolean, int and float.
	Type() ArgType

	// WithType sets the argument type.
	WithType(at ArgType) Arg

	// Optional specifies that an argument may be omitted. No non-optional arguments
	// should follow an optional one (no validation for this scenario as this is
	// the definition time exception, rather than incorrect input at runtime).
	Optional() bool

	// AsOptional sets the argument as optional.
	AsOptional() Arg

	AutoComplete() ArgAutoComplete

	WithAutoComplete(ac ArgAutoComplete) Arg
}

func DefaultArgAutoComplete(token string) []string {
	return []string{}
}

// NewArg creates a new positional argument.
func NewArg(key, descr string) Arg {
	return arg{
		key:   key,
		descr: descr,
		ac:    DefaultArgAutoComplete,
	}
}

type arg struct {
	key      string
	descr    string
	at       ArgType
	ac       ArgAutoComplete
	optional bool
}

func (a arg) Key() string {
	return a.key
}

func (a arg) Description() string {
	return a.descr
}

func (a arg) Type() ArgType {
	return a.at
}

func (a arg) Optional() bool {
	return a.optional
}

func (a arg) WithType(at ArgType) Arg {
	a.at = at
	return a
}

func (a arg) AsOptional() Arg {
	a.optional = true
	return a
}

func (a arg) AutoComplete() ArgAutoComplete {
	return a.ac
}

func (a arg) WithAutoComplete(ac ArgAutoComplete) Arg {
	a.ac = ac
	return a
}
