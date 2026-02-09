package core

import (
	"os"
	"strconv"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func checkTypeFileWhitespace(path string, relativePath string, expected string) error {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	warnTypeFileWhitespace(relativePath, expected, string(rawBytes))
	return nil
}

func warnTypeFileWhitespace(path string, expected string, raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == expected && raw != expected {
		zlog.Warn().
			Str("path", path).
			Str("found", strconv.QuoteToASCII(raw)).
			Str("expected", strconv.QuoteToASCII(expected)).
			Msg("malformed sentinel file")
	}
}
