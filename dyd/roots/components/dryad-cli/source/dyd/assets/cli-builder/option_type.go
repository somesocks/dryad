package cli_builder

// OptionType defines the type of permitted option values.
type OptionType int

// OptionType constants for string, boolean, int and number options and arguments.
const (
	OptionTypeString OptionType = iota
	OptionTypeBool
	OptionTypeInt
	OptionTypeNumber
	OptionTypeMultiString
	OptionTypeMultiBool
	OptionTypeMultiInt
	OptionTypeMultiNumber
)
