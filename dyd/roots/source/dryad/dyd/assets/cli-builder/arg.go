package cli_builder

type arg struct {
	key      string
	descr    string
	at       ArgType
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
