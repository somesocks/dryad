package core

type UnsafeHeapStemReference struct {
	BasePath     string
	Fingerprint  string
	Stems        *SafeHeapStemsReference
}

type SafeHeapStemReference struct {
	BasePath     string
	Fingerprint  string
	Stems        *SafeHeapStemsReference
}
