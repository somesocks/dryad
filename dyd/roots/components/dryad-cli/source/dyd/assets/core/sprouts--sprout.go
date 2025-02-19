package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (sprouts *SafeSproutsReference) Sprout(path string) (*UnsafeSproutReference) {

	// relative paths should be resolved,
	// relative to the base of the sprouts
	if !filepath.IsAbs(path) {
		path = filepath.Join(sprouts.BasePath, path)
	}

	// clean the resulting path
	path = filepath.Clean(path)
	
	var sproutRef = UnsafeSproutReference{
		BasePath: path,
		Sprouts: sprouts,
	}
	return &sproutRef
}