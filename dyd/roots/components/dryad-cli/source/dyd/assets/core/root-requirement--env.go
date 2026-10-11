package core

import (
	"dryad/internal/os"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/crypto/blake2b"
)

const RootRequirementEnvFingerprintQueryKey = "fingerprint"
const RootRequirementEnvDefaultQueryKey = "default"

var rootRequirementEnvNameRE = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func rootRequirementCanonicalEnvName(raw string) (error, string) {
	canonical := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(raw), "-", "_"), ".", "_"))
	if !rootRequirementEnvNameRE.MatchString(canonical) {
		return fmt.Errorf("invalid env requirement name %q; must canonicalize to [A-Z_][A-Z0-9_]*", raw), ""
	}

	return nil, canonical
}

func rootRequirementEnvValueFingerprint(value string) (error, string) {
	hash, err := blake2b.New(fingerprintDigestLen, []byte{})
	if err != nil {
		return err, ""
	}

	_, err = io.WriteString(hash, "env\u0000")
	if err != nil {
		return err, ""
	}

	_, err = io.WriteString(hash, value)
	if err != nil {
		return err, ""
	}

	hashInBytes := hash.Sum(nil)[:fingerprintDigestLen]
	return nil, fingerprintFormat(fingerprintVersionV2, fingerprintEncode(hashInBytes))
}

func rootRequirementEnvTargetString(envName string, fingerprint string, defaultValue *string) string {
	linkURL := url.URL{
		Scheme: "env",
		Opaque: envName,
	}
	query := url.Values{}
	if defaultValue != nil {
		query.Set(RootRequirementEnvDefaultQueryKey, *defaultValue)
	}
	if fingerprint != "" {
		query.Set(RootRequirementEnvFingerprintQueryKey, fingerprint)
	}
	linkURL.RawQuery = query.Encode()

	return linkURL.String()
}

type rootRequirementEnvTargetSpec struct {
	Name        string
	Fingerprint string
	Default     *string
}

func rootRequirementEnvValue(envName string, defaultValue *string) (string, bool) {
	value, exists := os.LookupEnv(envName)
	if !exists && defaultValue != nil {
		return *defaultValue, true
	}
	return value, exists
}

func rootRequirementEnvTargetFromURL(linkURL *url.URL) (error, rootRequirementEnvTargetSpec) {
	if linkURL.Fragment != "" {
		return fmt.Errorf("env requirement fragments are not supported"), rootRequirementEnvTargetSpec{}
	}

	linkPath := linkURL.Opaque
	if linkPath == "" {
		linkPath = linkURL.Path
	}
	if linkPath == "" {
		return fmt.Errorf("missing env requirement name"), rootRequirementEnvTargetSpec{}
	}

	err, envName := rootRequirementCanonicalEnvName(linkPath)
	if err != nil {
		return err, rootRequirementEnvTargetSpec{}
	}

	query, err := url.ParseQuery(linkURL.RawQuery)
	if err != nil {
		return err, rootRequirementEnvTargetSpec{}
	}
	var defaultValue *string
	if values, exists := query[RootRequirementEnvDefaultQueryKey]; exists {
		if len(values) != 1 {
			return fmt.Errorf("env requirement default must be specified once"), rootRequirementEnvTargetSpec{}
		}
		defaultValue = &values[0]
	}
	query.Del(RootRequirementEnvDefaultQueryKey)
	fingerprint := query.Get(RootRequirementEnvFingerprintQueryKey)
	if fingerprint != "" {
		err, _, _ = fingerprintParse(fingerprint)
		if err != nil {
			return err, rootRequirementEnvTargetSpec{}
		}
	}
	query.Del(RootRequirementEnvFingerprintQueryKey)
	if len(query) > 0 {
		return fmt.Errorf("unsupported env requirement query parameter"), rootRequirementEnvTargetSpec{}
	}

	return nil, rootRequirementEnvTargetSpec{
		Name:        envName,
		Fingerprint: fingerprint,
		Default:     defaultValue,
	}
}

func rootRequirementParseEnvTarget(raw string) (error, rootRequirementEnvTargetSpec, bool) {
	linkURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err, rootRequirementEnvTargetSpec{}, false
	}
	if linkURL.Scheme != "env" {
		return nil, rootRequirementEnvTargetSpec{}, false
	}

	err, envSpec := rootRequirementEnvTargetFromURL(linkURL)
	if err != nil {
		return err, rootRequirementEnvTargetSpec{}, true
	}

	return nil, envSpec, true
}

func RootRequirementEnvTargetNormalize(raw string) (error, string) {
	err, envSpec, isEnv := rootRequirementParseEnvTarget(raw)
	if err != nil {
		return err, ""
	}
	if !isEnv {
		return fmt.Errorf("env requirement target must use env scheme: %s", raw), ""
	}

	return nil, rootRequirementEnvTargetString(envSpec.Name, envSpec.Fingerprint, envSpec.Default)
}

func RootRequirementEnvTargetString(envName string, fingerprint string, defaultValue *string) string {
	return rootRequirementEnvTargetString(envName, fingerprint, defaultValue)
}
