package core

type UnsafeSproutsReference struct {
	BasePath string
	Garden *SafeGardenReference
}

type SafeSproutsReference struct {
	BasePath string
	Garden *SafeGardenReference
}

