package core

type UnsafeHeapSproutReference struct {
	BasePath     string
	Fingerprint  string
	Sprouts      *SafeHeapSproutsReference
}

type SafeHeapSproutReference struct {
	BasePath     string
	Fingerprint  string
	Sprouts      *SafeHeapSproutsReference
}
