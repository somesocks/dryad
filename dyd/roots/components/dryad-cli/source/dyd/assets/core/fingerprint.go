package core

import (
	"encoding/base32"
	"fmt"
	"strings"
)

const (
	fingerprintVersionV2 = "v2"
	fingerprintDigestLen = 16
)

var fingerprintEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func fingerprintEncode(digest []byte) string {
	return strings.ToLower(fingerprintEncoding.EncodeToString(digest))
}

func fingerprintFormat(version string, encoded string) string {
	return version + "-" + encoded
}

func fingerprintParse(fingerprint string) (error, string, string) {
	version, encoded, found := strings.Cut(strings.TrimSpace(fingerprint), "-")
	if !found || version == "" || encoded == "" {
		return fmt.Errorf("invalid fingerprint: %q", fingerprint), "", ""
	}
	if version != fingerprintVersionV2 {
		return fmt.Errorf("unsupported fingerprint version: %q", fingerprint), "", ""
	}
	if encoded != strings.ToLower(encoded) {
		return fmt.Errorf("fingerprint must use lowercase encoding: %q", fingerprint), "", ""
	}
	decoded, err := fingerprintEncoding.DecodeString(strings.ToUpper(encoded))
	if err != nil {
		return fmt.Errorf("invalid fingerprint encoding: %q", fingerprint), "", ""
	}
	if len(decoded) != fingerprintDigestLen {
		return fmt.Errorf("invalid fingerprint digest length: %q", fingerprint), "", ""
	}
	return nil, version, encoded
}
