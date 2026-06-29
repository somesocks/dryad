package core

import (
	"context"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type rootRequirementHTTPTargetSpec struct {
	SourceURL       string
	DestinationAs   string
	DestinationInto string
	Unpack          bool
	HasUnpack       bool
	ArchiveFormat   string
	Fingerprint     string
}

func rootRequirementHTTPMetadataValues(linkURL *url.URL) (error, url.Values) {
	if linkURL.Fragment == "" {
		return nil, url.Values{}
	}
	query, err := url.ParseQuery(linkURL.Fragment)
	if err != nil {
		return err, nil
	}
	return nil, query
}

func rootRequirementHTTPTargetFromURL(linkURL *url.URL) (error, rootRequirementHTTPTargetSpec) {
	if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
		return fmt.Errorf("http requirement target must use http or https scheme"), rootRequirementHTTPTargetSpec{}
	}
	if linkURL.Host == "" {
		return fmt.Errorf("http requirement target must specify a host"), rootRequirementHTTPTargetSpec{}
	}
	if linkURL.User != nil {
		return fmt.Errorf("http requirement target must not include userinfo"), rootRequirementHTTPTargetSpec{}
	}

	err, query := rootRequirementHTTPMetadataValues(linkURL)
	if err != nil {
		return err, rootRequirementHTTPTargetSpec{}
	}

	if _, ok := query[RootRequirementFileOptionalQueryKey]; ok {
		return fmt.Errorf("http requirement optional parameter is not supported"), rootRequirementHTTPTargetSpec{}
	}
	if _, ok := query["target"]; ok {
		return fmt.Errorf("http requirement target parameter is not supported; use as or into"), rootRequirementHTTPTargetSpec{}
	}

	destinationAsRaw, hasDestinationAs := query[RootRequirementFileAsQueryKey]
	destinationIntoRaw, hasDestinationInto := query[RootRequirementFileIntoQueryKey]
	if hasDestinationAs && hasDestinationInto {
		return fmt.Errorf("http requirement cannot specify both as and into"), rootRequirementHTTPTargetSpec{}
	}
	destinationAs := ""
	destinationInto := ""
	if hasDestinationAs {
		if len(destinationAsRaw) != 1 || destinationAsRaw[0] == "" {
			return fmt.Errorf("http requirement as path must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		var err error
		destinationAs, err = rootRequirementFilePlacementNormalize(destinationAsRaw[0], true)
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
	}
	if hasDestinationInto {
		if len(destinationIntoRaw) != 1 || destinationIntoRaw[0] == "" {
			return fmt.Errorf("http requirement into path must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		var err error
		destinationInto, err = rootRequirementFilePlacementNormalize(destinationIntoRaw[0], false)
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
	}
	query.Del(RootRequirementFileAsQueryKey)
	query.Del(RootRequirementFileIntoQueryKey)

	unpack := false
	unpackRaw, hasUnpack := query[RootRequirementFileUnpackQueryKey]
	if hasUnpack {
		if len(unpackRaw) != 1 || unpackRaw[0] == "" {
			return fmt.Errorf("http requirement unpack must be true or false"), rootRequirementHTTPTargetSpec{}
		}
		switch unpackRaw[0] {
		case "true":
			unpack = true
		case "false":
			unpack = false
		default:
			return fmt.Errorf("http requirement unpack must be true or false"), rootRequirementHTTPTargetSpec{}
		}
	}
	query.Del(RootRequirementFileUnpackQueryKey)

	archiveFormat := ""
	archiveFormatRaw, hasArchiveFormat := query[RootRequirementFileFormatQueryKey]
	if hasArchiveFormat {
		if len(archiveFormatRaw) != 1 || archiveFormatRaw[0] == "" {
			return fmt.Errorf("http requirement format must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		var err error
		archiveFormat, err = rootRequirementFileArchiveFormatNormalize(archiveFormatRaw[0])
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
	}
	query.Del(RootRequirementFileFormatQueryKey)

	fingerprintRaw, hasFingerprint := query[RootRequirementFileFingerprintQueryKey]
	fingerprint := ""
	if hasFingerprint {
		if len(fingerprintRaw) != 1 || fingerprintRaw[0] == "" {
			return fmt.Errorf("http requirement fingerprint must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		fingerprint = fingerprintRaw[0]
		if err, _, _ := fingerprintParse(fingerprint); err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
	}
	query.Del(RootRequirementFileFingerprintQueryKey)

	if len(query) > 0 {
		return fmt.Errorf("unsupported http requirement metadata parameter"), rootRequirementHTTPTargetSpec{}
	}

	sourceURL := *linkURL
	sourceURL.Fragment = ""
	sourceURL.RawFragment = ""

	return nil, rootRequirementHTTPTargetSpec{
		SourceURL:       sourceURL.String(),
		DestinationAs:   destinationAs,
		DestinationInto: destinationInto,
		Unpack:          unpack,
		HasUnpack:       hasUnpack,
		ArchiveFormat:   archiveFormat,
		Fingerprint:     fingerprint,
	}
}

func rootRequirementHTTPValidateStoredTargetSpec(httpSpec rootRequirementHTTPTargetSpec) error {
	if httpSpec.Fingerprint == "" {
		return fmt.Errorf("http requirement fingerprint is required")
	}
	if httpSpec.ArchiveFormat != "" && !httpSpec.Unpack {
		return fmt.Errorf("http requirement format requires unpack=true")
	}
	return nil
}

func rootRequirementHTTPTargetString(sourceURL string, destinationAs string, destinationInto string, unpack bool, archiveFormat string, fingerprint string) string {
	linkURL, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil {
		return sourceURL
	}
	linkURL.Fragment = ""
	linkURL.RawFragment = ""

	query := []string{}
	escapeQueryValue := func(value string) string {
		return strings.ReplaceAll(url.QueryEscape(value), "%2F", "/")
	}
	if destinationAs != "" {
		query = append(query, RootRequirementFileAsQueryKey+"="+escapeQueryValue(destinationAs))
	}
	if fingerprint != "" {
		query = append(query, RootRequirementFileFingerprintQueryKey+"="+escapeQueryValue(fingerprint))
	}
	if destinationInto != "" && destinationInto != "dyd/assets" {
		query = append(query, RootRequirementFileIntoQueryKey+"="+escapeQueryValue(destinationInto))
	}
	if unpack {
		query = append(query, RootRequirementFileUnpackQueryKey+"=true")
	}
	if archiveFormat != "" {
		query = append(query, RootRequirementFileFormatQueryKey+"="+escapeQueryValue(archiveFormat))
	}
	if len(query) > 0 {
		linkURL.Fragment = strings.Join(query, "&")
	}
	return linkURL.String()
}

func rootRequirementParseHTTPTarget(raw string) (error, rootRequirementHTTPTargetSpec, bool) {
	linkURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err, rootRequirementHTTPTargetSpec{}, false
	}
	if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
		return nil, rootRequirementHTTPTargetSpec{}, false
	}

	err, httpSpec := rootRequirementHTTPTargetFromURL(linkURL)
	if err != nil {
		return err, rootRequirementHTTPTargetSpec{}, true
	}

	return nil, httpSpec, true
}

func RootRequirementHTTPTargetNormalize(raw string) (error, string) {
	err, httpSpec, isHTTP := rootRequirementParseHTTPTarget(raw)
	if err != nil {
		return err, ""
	}
	if !isHTTP {
		return fmt.Errorf("http requirement target must use http or https scheme: %s", raw), ""
	}
	if err := rootRequirementHTTPValidateStoredTargetSpec(httpSpec); err != nil {
		return err, ""
	}

	return nil, rootRequirementHTTPTargetString(httpSpec.SourceURL, httpSpec.DestinationAs, httpSpec.DestinationInto, httpSpec.Unpack, httpSpec.ArchiveFormat, httpSpec.Fingerprint)
}

func RootRequirementHTTPTargetString(sourceURL string, destinationAs string, destinationInto string, unpack bool, archiveFormat string, fingerprint string) string {
	return rootRequirementHTTPTargetString(sourceURL, destinationAs, destinationInto, unpack, archiveFormat, fingerprint)
}

type RootRequirementHTTPLockTargetRequest struct {
	Garden             *SafeGardenReference
	Target             string
	DestinationAs      string
	HasDestinationAs   bool
	DestinationInto    string
	HasDestinationInto bool
	Unpack             bool
	HasUnpack          bool
	ArchiveFormat      string
	HasArchiveFormat   bool
	Fingerprint        string
	HasFingerprint     bool
}

func rootRequirementHTTPApplyLockOptions(httpSpec rootRequirementHTTPTargetSpec, req RootRequirementHTTPLockTargetRequest) (error, rootRequirementHTTPTargetSpec) {
	if req.HasDestinationAs {
		if req.DestinationAs == "" {
			return fmt.Errorf("http requirement as path must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		destinationAs, err := rootRequirementFilePlacementNormalize(req.DestinationAs, true)
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
		if httpSpec.DestinationAs != "" && httpSpec.DestinationAs != destinationAs {
			return fmt.Errorf("http requirement as specified both in target and --as"), rootRequirementHTTPTargetSpec{}
		}
		httpSpec.DestinationAs = destinationAs
	}
	if req.HasDestinationInto {
		if req.DestinationInto == "" {
			return fmt.Errorf("http requirement into path must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		destinationInto, err := rootRequirementFilePlacementNormalize(req.DestinationInto, false)
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
		if httpSpec.DestinationInto != "" && httpSpec.DestinationInto != destinationInto {
			return fmt.Errorf("http requirement into specified both in target and --into"), rootRequirementHTTPTargetSpec{}
		}
		httpSpec.DestinationInto = destinationInto
	}
	if httpSpec.DestinationAs != "" && httpSpec.DestinationInto != "" {
		return fmt.Errorf("http requirement cannot specify both as and into"), rootRequirementHTTPTargetSpec{}
	}
	if req.HasUnpack {
		if httpSpec.HasUnpack && httpSpec.Unpack != req.Unpack {
			return fmt.Errorf("http requirement unpack specified both in target and --unpack with different values"), rootRequirementHTTPTargetSpec{}
		}
		httpSpec.Unpack = req.Unpack
		httpSpec.HasUnpack = true
	}
	if req.HasArchiveFormat {
		if req.ArchiveFormat == "" {
			return fmt.Errorf("http requirement format must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		archiveFormat, err := rootRequirementFileArchiveFormatNormalize(req.ArchiveFormat)
		if err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
		if httpSpec.ArchiveFormat != "" && httpSpec.ArchiveFormat != archiveFormat {
			return fmt.Errorf("http requirement format specified both in target and --format"), rootRequirementHTTPTargetSpec{}
		}
		httpSpec.ArchiveFormat = archiveFormat
	}
	if httpSpec.ArchiveFormat != "" && !httpSpec.Unpack {
		return fmt.Errorf("http requirement format requires unpack=true"), rootRequirementHTTPTargetSpec{}
	}
	if req.HasFingerprint {
		if req.Fingerprint == "" {
			return fmt.Errorf("http requirement fingerprint must not be empty"), rootRequirementHTTPTargetSpec{}
		}
		if err, _, _ := fingerprintParse(req.Fingerprint); err != nil {
			return err, rootRequirementHTTPTargetSpec{}
		}
		if httpSpec.Fingerprint != "" && httpSpec.Fingerprint != req.Fingerprint {
			return fmt.Errorf("http requirement fingerprint specified both in target and --fingerprint"), rootRequirementHTTPTargetSpec{}
		}
		httpSpec.Fingerprint = req.Fingerprint
	}

	return nil, httpSpec
}

func RootRequirementHTTPLockTarget(ctx *task.ExecutionContext, req RootRequirementHTTPLockTargetRequest) (error, string) {
	err, httpSpec, isHTTP := rootRequirementParseHTTPTarget(req.Target)
	if err != nil {
		return err, ""
	}
	if !isHTTP {
		return fmt.Errorf("http requirement target must use http or https scheme: %s", req.Target), ""
	}

	err, httpSpec = rootRequirementHTTPApplyLockOptions(httpSpec, req)
	if err != nil {
		return err, ""
	}

	if httpSpec.Fingerprint == "" {
		err, stem := rootRequirementHTTPDownloadBuildStem(ctx, RootRequirementHTTPBuildStemRequest{
			Garden:          req.Garden,
			SourceURL:       httpSpec.SourceURL,
			DestinationAs:   httpSpec.DestinationAs,
			DestinationInto: httpSpec.DestinationInto,
			Unpack:          httpSpec.Unpack,
			ArchiveFormat:   httpSpec.ArchiveFormat,
		})
		if err != nil {
			return err, ""
		}
		if stem == nil {
			return fmt.Errorf("missing http dependency build result: %s", httpSpec.SourceURL), ""
		}
		httpSpec.Fingerprint = stem.Fingerprint
	}

	return nil, rootRequirementHTTPTargetString(httpSpec.SourceURL, httpSpec.DestinationAs, httpSpec.DestinationInto, httpSpec.Unpack, httpSpec.ArchiveFormat, httpSpec.Fingerprint)
}

type rootRequirementHTTPAuthCredential struct {
	EnvName string
	Literal string
}

type rootRequirementHTTPAuthSpec struct {
	Scheme      string
	Credentials []rootRequirementHTTPAuthCredential
}

type rootRequirementHTTPRemoteConfig struct {
	VHost string
	Host  string
	Auth  rootRequirementHTTPAuthSpec
}

type rootRequirementHTTPRemoteConfigRequest struct {
	Garden *SafeGardenReference
	VHost  string
}

func rootRequirementHTTPRemoteConfigPath(garden *SafeGardenReference, vhost string, name string) string {
	return filepath.Join(garden.BasePath, "dyd", "shed", "remotes", vhost, name)
}

func rootRequirementHTTPReadOptionalTrimmedFile(path string) (error, string, bool) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", false
		}
		return err, "", false
	}
	return nil, strings.TrimSpace(string(bytes)), true
}

func rootRequirementHTTPValidateRemoteVHost(vhost string) error {
	if vhost == "" {
		return fmt.Errorf("remote vhost must not be empty")
	}
	if strings.TrimSpace(vhost) != vhost || vhost == "." || vhost == ".." {
		return fmt.Errorf("invalid remote vhost: %s", vhost)
	}
	if strings.Contains(vhost, "://") || strings.ContainsAny(vhost, " /\\\t\n\r?#@\x00") {
		return fmt.Errorf("invalid remote vhost: %s", vhost)
	}
	for _, r := range vhost {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("invalid remote vhost: %s", vhost)
		}
	}
	return nil
}

func rootRequirementHTTPValidateRemoteHost(host string) error {
	if host == "" {
		return fmt.Errorf("remote host must not be empty")
	}
	if strings.TrimSpace(host) != host || strings.ContainsAny(host, " \\\t\n\r/?#@\x00") || strings.Contains(host, "://") {
		return fmt.Errorf("invalid remote host: %s", host)
	}
	for _, r := range host {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("invalid remote host: %s", host)
		}
	}
	if host == "." || host == ".." {
		return fmt.Errorf("invalid remote host: %s", host)
	}
	return nil
}

func rootRequirementHTTPAuthCredentialParse(raw string) (error, rootRequirementHTTPAuthCredential) {
	if raw == "" {
		return fmt.Errorf("auth credential must not be empty"), rootRequirementHTTPAuthCredential{}
	}
	if strings.HasPrefix(raw, "env:") {
		envName := strings.TrimPrefix(raw, "env:")
		if !rootRequirementEnvNameRE.MatchString(envName) {
			return fmt.Errorf("invalid auth env credential %q", raw), rootRequirementHTTPAuthCredential{}
		}
		return nil, rootRequirementHTTPAuthCredential{EnvName: envName}
	}
	return nil, rootRequirementHTTPAuthCredential{Literal: raw}
}

func rootRequirementHTTPAuthParse(raw string) (error, rootRequirementHTTPAuthSpec) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return nil, rootRequirementHTTPAuthSpec{Scheme: "none"}
	}

	spec := rootRequirementHTTPAuthSpec{Scheme: fields[0]}
	switch spec.Scheme {
	case "none":
		if len(fields) != 1 {
			return fmt.Errorf("auth none takes no credentials"), rootRequirementHTTPAuthSpec{}
		}
	case "bearer":
		if len(fields) != 2 {
			return fmt.Errorf("auth bearer takes one credential"), rootRequirementHTTPAuthSpec{}
		}
		err, credential := rootRequirementHTTPAuthCredentialParse(fields[1])
		if err != nil {
			return err, rootRequirementHTTPAuthSpec{}
		}
		spec.Credentials = []rootRequirementHTTPAuthCredential{credential}
	case "basic":
		if len(fields) != 3 {
			return fmt.Errorf("auth basic takes two credentials"), rootRequirementHTTPAuthSpec{}
		}
		err, username := rootRequirementHTTPAuthCredentialParse(fields[1])
		if err != nil {
			return err, rootRequirementHTTPAuthSpec{}
		}
		err, password := rootRequirementHTTPAuthCredentialParse(fields[2])
		if err != nil {
			return err, rootRequirementHTTPAuthSpec{}
		}
		spec.Credentials = []rootRequirementHTTPAuthCredential{username, password}
	default:
		return fmt.Errorf("unsupported http remote auth scheme: %s", spec.Scheme), rootRequirementHTTPAuthSpec{}
	}

	return nil, spec
}

func rootRequirementHTTPResolveRemoteConfig(ctx *task.ExecutionContext, req rootRequirementHTTPRemoteConfigRequest) (error, rootRequirementHTTPRemoteConfig) {
	if err := rootRequirementHTTPValidateRemoteVHost(req.VHost); err != nil {
		return err, rootRequirementHTTPRemoteConfig{}
	}

	host := req.VHost
	err, hostRaw, hasHost := rootRequirementHTTPReadOptionalTrimmedFile(rootRequirementHTTPRemoteConfigPath(req.Garden, req.VHost, "host"))
	if err != nil {
		return err, rootRequirementHTTPRemoteConfig{}
	}
	if hasHost {
		host = hostRaw
	}
	if err := rootRequirementHTTPValidateRemoteHost(host); err != nil {
		return err, rootRequirementHTTPRemoteConfig{}
	}

	err, authRaw, _ := rootRequirementHTTPReadOptionalTrimmedFile(rootRequirementHTTPRemoteConfigPath(req.Garden, req.VHost, "auth"))
	if err != nil {
		return err, rootRequirementHTTPRemoteConfig{}
	}
	err, auth := rootRequirementHTTPAuthParse(authRaw)
	if err != nil {
		return err, rootRequirementHTTPRemoteConfig{}
	}

	return nil, rootRequirementHTTPRemoteConfig{
		VHost: req.VHost,
		Host:  host,
		Auth:  auth,
	}
}

var memoRootRequirementHTTPResolveRemoteConfig = task.Memoize(
	rootRequirementHTTPResolveRemoteConfig,
	func(ctx *task.ExecutionContext, req rootRequirementHTTPRemoteConfigRequest) (error, any) {
		type Key struct {
			Group      string
			GardenPath string
			VHost      string
		}
		gardenPath := ""
		if req.Garden != nil {
			gardenPath = req.Garden.BasePath
		}
		return nil, Key{
			Group:      "RootRequirementHTTP.RemoteConfig.Resolve",
			GardenPath: gardenPath,
			VHost:      req.VHost,
		}
	},
)

func rootRequirementHTTPAuthCredentialValue(credential rootRequirementHTTPAuthCredential) (error, string) {
	if credential.EnvName != "" {
		value, ok := os.LookupEnv(credential.EnvName)
		if !ok {
			return fmt.Errorf("missing auth env credential: %s", credential.EnvName), ""
		}
		return nil, value
	}
	return nil, credential.Literal
}

func rootRequirementHTTPApplyAuth(req *http.Request, auth rootRequirementHTTPAuthSpec) error {
	switch auth.Scheme {
	case "", "none":
		return nil
	case "bearer":
		err, token := rootRequirementHTTPAuthCredentialValue(auth.Credentials[0])
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case "basic":
		err, username := rootRequirementHTTPAuthCredentialValue(auth.Credentials[0])
		if err != nil {
			return err
		}
		err, password := rootRequirementHTTPAuthCredentialValue(auth.Credentials[1])
		if err != nil {
			return err
		}
		req.SetBasicAuth(username, password)
	default:
		return fmt.Errorf("unsupported http remote auth scheme: %s", auth.Scheme)
	}
	return nil
}

type RootRequirementHTTPBuildStemRequest struct {
	Garden          *SafeGardenReference
	SourceURL       string
	DestinationAs   string
	DestinationInto string
	Unpack          bool
	ArchiveFormat   string
	Fingerprint     string
}

func rootRequirementHTTPDownloadBuildStem(ctx *task.ExecutionContext, req RootRequirementHTTPBuildStemRequest) (error, *SafeHeapStemReference) {
	sourceURL, err := url.Parse(req.SourceURL)
	if err != nil {
		return err, nil
	}
	if sourceURL.Scheme != "http" && sourceURL.Scheme != "https" {
		return fmt.Errorf("http requirement target must use http or https scheme"), nil
	}
	if sourceURL.User != nil {
		return fmt.Errorf("http requirement target must not include userinfo"), nil
	}

	err, remoteConfig := memoRootRequirementHTTPResolveRemoteConfig(ctx, rootRequirementHTTPRemoteConfigRequest{
		Garden: req.Garden,
		VHost:  sourceURL.Host,
	})
	if err != nil {
		return err, nil
	}
	sourceURL.Host = remoteConfig.Host

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sourceURL.String(), nil)
	if err != nil {
		return err, nil
	}
	if err := rootRequirementHTTPApplyAuth(httpReq, remoteConfig.Auth); err != nil {
		return err, nil
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http requirement fetch failed for %s: %w", req.SourceURL, err), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http requirement fetch failed for %s: %s", req.SourceURL, resp.Status), nil
	}

	downloadDir, err := os.MkdirTemp("", "dryad-http-*")
	if err != nil {
		return err, nil
	}
	defer os.RemoveAll(downloadDir)
	downloadName := filepath.Base(sourceURL.Path)
	if downloadName == "." || downloadName == "" || downloadName == "/" {
		downloadName = "download"
	}
	downloadPath := filepath.Join(downloadDir, downloadName)
	downloadFile, err := os.OpenFile(downloadPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err, nil
	}
	_, copyErr := io.Copy(downloadFile, resp.Body)
	closeErr := downloadFile.Close()
	if copyErr != nil {
		return copyErr, nil
	}
	if closeErr != nil {
		return closeErr, nil
	}

	err, stem := RootRequirementFileBuildStem(ctx, RootRequirementFileBuildStemRequest{
		Garden:          req.Garden,
		SourcePath:      downloadPath,
		DestinationAs:   req.DestinationAs,
		DestinationInto: req.DestinationInto,
		Unpack:          req.Unpack,
		ArchiveFormat:   req.ArchiveFormat,
	})
	if err != nil {
		return err, nil
	}
	if stem == nil {
		return fmt.Errorf("missing http dependency build result: %s", req.SourceURL), nil
	}
	return nil, stem
}

func RootRequirementHTTPBuildStem(ctx *task.ExecutionContext, req RootRequirementHTTPBuildStemRequest) (error, *SafeHeapStemReference) {
	if req.Fingerprint == "" {
		return fmt.Errorf("http requirement fingerprint is required"), nil
	}
	if err, _, _ := fingerprintParse(req.Fingerprint); err != nil {
		return err, nil
	}

	err, heap := req.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, nil
	}
	err, stems := heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}
	err, cachedStem := stems.Stem(req.Fingerprint).Resolve(ctx)
	if err == nil {
		return nil, cachedStem
	}

	err, stem := rootRequirementHTTPDownloadBuildStem(ctx, req)
	if err != nil {
		return err, nil
	}
	if stem == nil {
		return fmt.Errorf("missing http dependency build result: %s", req.SourceURL), nil
	}
	if stem.Fingerprint != req.Fingerprint {
		return fmt.Errorf("http requirement fingerprint mismatch for %s: expected %s, got %s", req.SourceURL, req.Fingerprint, stem.Fingerprint), nil
	}
	return nil, stem
}
