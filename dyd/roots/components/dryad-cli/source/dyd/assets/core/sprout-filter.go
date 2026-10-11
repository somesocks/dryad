package core

import "context"

type SproutFilter func(context.Context, *SafeSproutReference) (error, bool)
