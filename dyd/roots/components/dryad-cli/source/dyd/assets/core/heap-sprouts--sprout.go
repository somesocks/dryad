package core

import (
	"strings"
)

func (sprouts *SafeHeapSproutsReference) Sprout(fingerprint string) *UnsafeHeapSproutReference {
	fingerprint = strings.TrimSpace(fingerprint)
	var heapSproutRef = UnsafeHeapSproutReference{
		BasePath:    "",
		Fingerprint: fingerprint,
		Sprouts:     sprouts,
	}
	return &heapSproutRef
}
