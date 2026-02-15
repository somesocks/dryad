package core

type UnsafeHeapSproutReference struct {
	BasePath string
	Sprouts  *SafeHeapSproutsReference
}

type SafeHeapSproutReference struct {
	BasePath string
	Sprouts  *SafeHeapSproutsReference
}
