package cli

// ArgType defines the type of permitted arg values.
type ArgType int

// ArgType constants for string, boolean, int and number options and arguments.
const (
	ArgTypeString ArgType = iota
	ArgTypeBool
	ArgTypeInt
	ArgTypeNumber
)
