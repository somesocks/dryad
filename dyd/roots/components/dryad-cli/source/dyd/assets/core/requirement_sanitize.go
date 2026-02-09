package core

import (
	"strconv"

	zlog "github.com/rs/zerolog/log"
)

func warnRequirementFileWhitespace(path string, raw string, expected string) {
	if raw != expected {
		zlog.Warn().
			Str("path", path).
			Str("found", strconv.QuoteToASCII(raw)).
			Str("expected", strconv.QuoteToASCII(expected)).
			Msg("malformed requirement file")
	}
}
