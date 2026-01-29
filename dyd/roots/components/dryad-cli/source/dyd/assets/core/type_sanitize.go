package core

import (
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func sanitizeTypeFile(path string, expected string) error {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	raw := string(rawBytes)
	trimmed := strings.TrimSpace(raw)

	if raw != expected {
		zlog.Warn().
			Str("path", path).
			Str("raw", strings.ReplaceAll(raw, "\n", "\\n")).
			Str("expected", expected).
			Msg("type file is incorrect")
	}

	// Only sanitize whitespace mistakes. Do not rewrite to an unexpected type.
	if trimmed == expected && raw != expected {
		return os.WriteFile(path, []byte(expected), os.ModePerm)
	}

	return nil
}

func warnTypeFileWhitespace(path string, expected string, raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == expected && raw != expected {
		zlog.Warn().
			Str("path", path).
			Str("raw", strings.ReplaceAll(raw, "\n", "\\n")).
			Str("expected", expected).
			Msg("type file contains whitespace")
	}
}
