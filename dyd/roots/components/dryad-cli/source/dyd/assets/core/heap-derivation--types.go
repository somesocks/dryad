package core

type UnsafeHeapDerivationReference struct {
	BasePath           string
	SourceFingerprint  string
	ResultFingerprint  string
	Derivations        *SafeHeapDerivationsReference
}

type SafeHeapDerivationReference struct {
	BasePath           string
	SourceFingerprint  string
	ResultFingerprint  string
	Source             *UnsafeHeapStemReference
	Result             *UnsafeHeapStemReference
	Derivations        *SafeHeapDerivationsReference
}
