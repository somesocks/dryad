package core

import (
	"strings"
)

func (stems *SafeHeapStemsReference) Stem(fingerprint string) *UnsafeHeapStemReference {
	fingerprint = strings.TrimSpace(fingerprint)
	var heapStemRef = UnsafeHeapStemReference{
		BasePath:    "",
		Fingerprint: fingerprint,
		Stems:       stems,
	}
	return &heapStemRef
}
