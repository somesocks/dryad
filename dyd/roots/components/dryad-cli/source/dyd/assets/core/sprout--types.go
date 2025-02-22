

package core

type UnsafeSproutReference struct {
	BasePath string
	Sprouts *SafeSproutsReference
}

type SafeSproutReference struct {
	BasePath string
	Sprouts *SafeSproutsReference
}

