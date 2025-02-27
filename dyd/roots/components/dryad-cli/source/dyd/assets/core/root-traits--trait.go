package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (traits *SafeRootTraitsReference) Trait(path string) (*UnsafeRootTraitReference) {
	var rootTraitRef = UnsafeRootTraitReference{
		BasePath: filepath.Join(traits.BasePath, path),
		Traits: traits,
	}
	return &rootTraitRef
}