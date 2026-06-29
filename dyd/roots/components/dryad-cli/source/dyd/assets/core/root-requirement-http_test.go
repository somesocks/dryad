package core

import (
	"archive/tar"
	"bytes"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeGardenForHTTPTest(t *testing.T) string {
	t.Helper()
	gardenPath := t.TempDir()
	makeWritableForCleanupForTest(t, gardenPath)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "type"), "garden")
	return gardenPath
}

func buildExpectedHTTPStemForTest(t *testing.T, content string, destinationAs string) string {
	t.Helper()
	gardenPath := makeGardenForHTTPTest(t)
	sourcePath := filepath.Join(t.TempDir(), "data.txt")
	writeFileForTest(t, sourcePath, content)
	err, stem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourcePath:    sourcePath,
		DestinationAs: destinationAs,
	})
	assert.Nil(t, err)
	assert.NotNil(t, stem)
	return stem.Fingerprint
}

func makeTarArchiveForHTTPTest(t *testing.T, name string, contents string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)
	assert.Nil(t, tarWriter.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(contents)),
	}))
	_, err := tarWriter.Write([]byte(contents))
	assert.Nil(t, err)
	assert.Nil(t, tarWriter.Close())
	return buf.Bytes()
}

func TestRootRequirementHTTPTargetNormalize(t *testing.T) {
	assert := assert.New(t)

	err, target := RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz?download=1#unpack=true&into=dyd/assets/vendor&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Nil(err)
	assert.Equal("https://example.com/pkg.tar.gz?download=1#fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa&into=dyd/assets/vendor&unpack=true", target)

	err, target = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.bz2?download=1#unpack=true&format=tbz&into=dyd/assets/vendor&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Nil(err)
	assert.Equal("https://example.com/pkg.tar.bz2?download=1#fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa&into=dyd/assets/vendor&unpack=true&format=tar.bz2", target)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz?download=1")
	assert.NotNil(err)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz#unknown=value&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.NotNil(err)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz#optional=true&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.NotNil(err)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz#as=dyd/commands/pkg&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.NotNil(err)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz#format=tar&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.NotNil(err)

	err, _ = RootRequirementHTTPTargetNormalize("https://example.com/pkg.tar.gz#unpack=true&format=rar&fingerprint=v2-aaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.NotNil(err)
}

func TestRootRequirementHTTPAuthParse(t *testing.T) {
	assert := assert.New(t)

	err, auth := rootRequirementHTTPAuthParse("bearer env:GITHUB_TOKEN")
	assert.Nil(err)
	assert.Equal("bearer", auth.Scheme)
	assert.Equal("GITHUB_TOKEN", auth.Credentials[0].EnvName)

	err, auth = rootRequirementHTTPAuthParse("basic user password")
	assert.Nil(err)
	assert.Equal("basic", auth.Scheme)
	assert.Equal("user", auth.Credentials[0].Literal)
	assert.Equal("password", auth.Credentials[1].Literal)

	err, _ = rootRequirementHTTPAuthParse("bearer env:not-valid")
	assert.NotNil(err)

	err, _ = rootRequirementHTTPAuthParse("bearer a b")
	assert.NotNil(err)
}

func TestRootRequirementHTTPRemoteConfigMemoizesWithinContext(t *testing.T) {
	assert := assert.New(t)
	gardenPath := makeGardenForHTTPTest(t)
	remotePath := filepath.Join(gardenPath, "dyd", "shed", "remotes", "repo")
	writeFileForTest(t, filepath.Join(remotePath, "host"), "first.example.com")
	writeFileForTest(t, filepath.Join(remotePath, "auth"), "bearer first-token")
	garden := &SafeGardenReference{BasePath: gardenPath}
	ctx := task.NewContext(1)

	err, config := memoRootRequirementHTTPResolveRemoteConfig(ctx, rootRequirementHTTPRemoteConfigRequest{Garden: garden, VHost: "repo"})
	assert.Nil(err)
	assert.Equal("first.example.com", config.Host)
	assert.Equal("first-token", config.Auth.Credentials[0].Literal)

	writeFileForTest(t, filepath.Join(remotePath, "host"), "second.example.com")
	err, config = memoRootRequirementHTTPResolveRemoteConfig(ctx, rootRequirementHTTPRemoteConfigRequest{Garden: garden, VHost: "repo"})
	assert.Nil(err)
	assert.Equal("first.example.com", config.Host)

	err, config = memoRootRequirementHTTPResolveRemoteConfig(task.NewContext(1), rootRequirementHTTPRemoteConfigRequest{Garden: garden, VHost: "repo"})
	assert.Nil(err)
	assert.Equal("second.example.com", config.Host)
}

func TestRootRequirementHTTPRemoteConfigRejectsUnsafeVHost(t *testing.T) {
	assert := assert.New(t)
	gardenPath := makeGardenForHTTPTest(t)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "shed", "host"), "escaped.example.com")
	garden := &SafeGardenReference{BasePath: gardenPath}

	invalidVHosts := []string{
		".",
		"..",
		"../escape",
		"escape/../host",
		`escape\host`,
		"bad host",
		"host?query",
		"host#fragment",
		"host@user",
		"http://host",
		"host\x00name",
	}
	for _, vhost := range invalidVHosts {
		err, _ := rootRequirementHTTPResolveRemoteConfig(task.NewContext(1), rootRequirementHTTPRemoteConfigRequest{Garden: garden, VHost: vhost})
		assert.NotNil(err, "expected invalid vhost %q to fail", vhost)
	}

	err, config := rootRequirementHTTPResolveRemoteConfig(task.NewContext(1), rootRequirementHTTPRemoteConfigRequest{Garden: garden, VHost: "repo:443"})
	assert.Nil(err)
	assert.Equal("repo:443", config.Host)
}

func TestRootRequirementHTTPBuildStem_DownloadsWithRemoteAuth(t *testing.T) {
	assert := assert.New(t)
	expectedFingerprint := buildExpectedHTTPStemForTest(t, "remote-data", "dyd/assets/data.txt")
	gardenPath := makeGardenForHTTPTest(t)
	t.Setenv("DRYAD_HTTP_TOKEN", "secret-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal("download=1", r.URL.RawQuery)
		fmt.Fprint(w, "remote-data")
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	assert.Nil(err)

	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "shed", "remotes", "artifact", "host"), serverURL.Host)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "shed", "remotes", "artifact", "auth"), "bearer env:DRYAD_HTTP_TOKEN")

	err, stem := RootRequirementHTTPBuildStem(task.NewContext(1), RootRequirementHTTPBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourceURL:     "http://artifact/data.txt?download=1",
		DestinationAs: "dyd/assets/data.txt",
		Fingerprint:   expectedFingerprint,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal(expectedFingerprint, stem.Fingerprint)
	assert.Equal("remote-data", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "data.txt")))
}

func TestRootRequirementHTTPLockTarget_FetchesMissingFingerprint(t *testing.T) {
	assert := assert.New(t)
	expectedFingerprint := buildExpectedHTTPStemForTest(t, "locked-data", "dyd/assets/data.txt")
	gardenPath := makeGardenForHTTPTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("download=1", r.URL.RawQuery)
		fmt.Fprint(w, "locked-data")
	}))
	defer server.Close()

	err, target := RootRequirementHTTPLockTarget(task.NewContext(1), RootRequirementHTTPLockTargetRequest{
		Garden:           &SafeGardenReference{BasePath: gardenPath},
		Target:           server.URL + "/data.txt?download=1",
		DestinationAs:    "dyd/assets/data.txt",
		HasDestinationAs: true,
	})
	assert.Nil(err)
	assert.Equal(server.URL+"/data.txt?download=1#as=dyd/assets/data.txt&fingerprint="+expectedFingerprint, target)

	err, normalized := RootRequirementHTTPTargetNormalize(target)
	assert.Nil(err)
	assert.Equal(target, normalized)
}

func TestRootRequirementHTTPLockTarget_UsesArchiveFormatOverride(t *testing.T) {
	assert := assert.New(t)
	gardenPath := makeGardenForHTTPTest(t)
	archive := makeTarArchiveForHTTPTest(t, "contents/value.txt", "packed")
	archivePath := filepath.Join(t.TempDir(), "pkg.tar.bz2")
	writeFileForTest(t, archivePath, string(archive))
	err, expectedStem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourcePath:      archivePath,
		DestinationInto: "dyd/assets/vendor",
		Unpack:          true,
		ArchiveFormat:   RootRequirementFileArchiveFormatTar,
	})
	assert.Nil(err)
	assert.NotNil(expectedStem)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archive)
	}))
	defer server.Close()

	err, target := RootRequirementHTTPLockTarget(task.NewContext(1), RootRequirementHTTPLockTargetRequest{
		Garden:             &SafeGardenReference{BasePath: gardenPath},
		Target:             server.URL + "/pkg.tar.bz2",
		DestinationInto:    "dyd/assets/vendor",
		HasDestinationInto: true,
		Unpack:             true,
		HasUnpack:          true,
		ArchiveFormat:      "tar",
		HasArchiveFormat:   true,
	})
	assert.Nil(err)
	assert.Equal(server.URL+"/pkg.tar.bz2#fingerprint="+expectedStem.Fingerprint+"&into=dyd/assets/vendor&unpack=true&format=tar", target)

	err, httpSpec, isHTTP := rootRequirementParseHTTPTarget(target)
	assert.Nil(err)
	assert.True(isHTTP)
	err, stem := RootRequirementHTTPBuildStem(task.NewContext(1), RootRequirementHTTPBuildStemRequest{
		Garden:          &SafeGardenReference{BasePath: gardenPath},
		SourceURL:       httpSpec.SourceURL,
		DestinationInto: httpSpec.DestinationInto,
		Unpack:          httpSpec.Unpack,
		ArchiveFormat:   httpSpec.ArchiveFormat,
		Fingerprint:     httpSpec.Fingerprint,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal("packed", readTrimmedFileForTest(t, filepath.Join(stem.BasePath, "dyd", "assets", "vendor", "pkg", "contents", "value.txt")))
}

func TestRootRequirementHTTPBuildStem_CacheHitSkipsNetworkAndAuth(t *testing.T) {
	assert := assert.New(t)
	gardenPath := makeGardenForHTTPTest(t)
	sourcePath := filepath.Join(t.TempDir(), "data.txt")
	writeFileForTest(t, sourcePath, "cached-data")
	err, cachedStem := RootRequirementFileBuildStem(task.NewContext(1), RootRequirementFileBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourcePath:    sourcePath,
		DestinationAs: "dyd/assets/data.txt",
	})
	assert.Nil(err)
	assert.NotNil(cachedStem)

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		fmt.Fprint(w, "unexpected")
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	assert.Nil(err)
	writeFileForTest(t, filepath.Join(gardenPath, "dyd", "shed", "remotes", serverURL.Host, "auth"), "bearer env:MISSING_HTTP_TOKEN")

	err, stem := RootRequirementHTTPBuildStem(task.NewContext(1), RootRequirementHTTPBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourceURL:     server.URL + "/data.txt",
		DestinationAs: "dyd/assets/data.txt",
		Fingerprint:   cachedStem.Fingerprint,
	})
	assert.Nil(err)
	assert.NotNil(stem)
	assert.Equal(cachedStem.Fingerprint, stem.Fingerprint)
	assert.Equal(int32(0), hits.Load())
}

func TestRootRequirementHTTPBuildStem_FingerprintMismatchFails(t *testing.T) {
	assert := assert.New(t)
	gardenPath := makeGardenForHTTPTest(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "wrong-data")
	}))
	defer server.Close()

	err, stem := RootRequirementHTTPBuildStem(task.NewContext(1), RootRequirementHTTPBuildStemRequest{
		Garden:        &SafeGardenReference{BasePath: gardenPath},
		SourceURL:     server.URL + "/data.txt",
		DestinationAs: "dyd/assets/data.txt",
		Fingerprint:   "v2-aaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	assert.NotNil(err)
	assert.Nil(stem)
}
