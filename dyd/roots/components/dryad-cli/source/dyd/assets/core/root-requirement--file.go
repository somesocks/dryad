package core

import (
	"archive/tar"
	"compress/gzip"
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const (
	RootRequirementFileFingerprintQueryKey = "fingerprint"
	RootRequirementFileAsQueryKey          = "as"
	RootRequirementFileIntoQueryKey        = "into"
	RootRequirementFileOptionalQueryKey    = "optional"
	RootRequirementFileUnpackQueryKey      = "unpack"
)

type rootRequirementFileTargetSpec struct {
	SourcePath      string
	DestinationAs   string
	DestinationInto string
	Optional        bool
	Unpack          bool
	Fingerprint     string
}

func rootRequirementFilePlacementDefaultInto(destinationInto string) string {
	if destinationInto == "" {
		return "dyd/assets"
	}
	return destinationInto
}

func rootRequirementFilePlacementNormalize(path string, exact bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file requirement placement path must not be empty")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("file requirement placement path must be relative: %s", path)
	}

	cleanPath := filepath.ToSlash(filepath.Clean(path))
	if cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return "", fmt.Errorf("file requirement placement path escapes package: %s", path)
	}

	for _, prefix := range []string{"dyd/assets", "dyd/docs", "dyd/traits", "dyd/secrets"} {
		if cleanPath == prefix {
			if exact {
				return "", fmt.Errorf("file requirement as path must be below %s", prefix)
			}
			return cleanPath, nil
		}
		if strings.HasPrefix(cleanPath, prefix+"/") {
			return cleanPath, nil
		}
	}

	return "", fmt.Errorf("file requirement placement path must be under dyd/assets, dyd/docs, dyd/traits, or dyd/secrets: %s", path)
}

func rootRequirementFileTargetFromURL(linkURL *url.URL) (error, rootRequirementFileTargetSpec) {
	if linkURL.Fragment != "" {
		return fmt.Errorf("file requirement fragments are not supported"), rootRequirementFileTargetSpec{}
	}
	if linkURL.Host != "" {
		return fmt.Errorf("file requirement host is not supported: %s", linkURL.Host), rootRequirementFileTargetSpec{}
	}

	linkPath := linkURL.Opaque
	if linkPath == "" {
		linkPath = linkURL.Path
	}
	if linkPath == "" {
		return fmt.Errorf("missing file requirement path"), rootRequirementFileTargetSpec{}
	}
	if filepath.IsAbs(linkPath) {
		return fmt.Errorf("file requirement path must be relative: %s", linkPath), rootRequirementFileTargetSpec{}
	}

	query := linkURL.Query()
	if _, ok := query["target"]; ok {
		return fmt.Errorf("file requirement target parameter is not supported; use as or into"), rootRequirementFileTargetSpec{}
	}

	destinationAsRaw, hasDestinationAs := query[RootRequirementFileAsQueryKey]
	destinationIntoRaw, hasDestinationInto := query[RootRequirementFileIntoQueryKey]
	if hasDestinationAs && hasDestinationInto {
		return fmt.Errorf("file requirement cannot specify both as and into"), rootRequirementFileTargetSpec{}
	}
	destinationAs := ""
	destinationInto := ""
	if hasDestinationAs {
		if len(destinationAsRaw) != 1 || destinationAsRaw[0] == "" {
			return fmt.Errorf("file requirement as path must not be empty"), rootRequirementFileTargetSpec{}
		}
		var err error
		destinationAs, err = rootRequirementFilePlacementNormalize(destinationAsRaw[0], true)
		if err != nil {
			return err, rootRequirementFileTargetSpec{}
		}
	}
	if hasDestinationInto {
		if len(destinationIntoRaw) != 1 || destinationIntoRaw[0] == "" {
			return fmt.Errorf("file requirement into path must not be empty"), rootRequirementFileTargetSpec{}
		}
		var err error
		destinationInto, err = rootRequirementFilePlacementNormalize(destinationIntoRaw[0], false)
		if err != nil {
			return err, rootRequirementFileTargetSpec{}
		}
	}
	query.Del(RootRequirementFileAsQueryKey)
	query.Del(RootRequirementFileIntoQueryKey)

	optional := false
	optionalRaw := query.Get(RootRequirementFileOptionalQueryKey)
	if optionalRaw != "" {
		switch optionalRaw {
		case "true":
			optional = true
		case "false":
			optional = false
		default:
			return fmt.Errorf("file requirement optional must be true or false"), rootRequirementFileTargetSpec{}
		}
	}
	query.Del(RootRequirementFileOptionalQueryKey)

	unpack := false
	unpackRaw := query.Get(RootRequirementFileUnpackQueryKey)
	if unpackRaw != "" {
		switch unpackRaw {
		case "true":
			unpack = true
		case "false":
			unpack = false
		default:
			return fmt.Errorf("file requirement unpack must be true or false"), rootRequirementFileTargetSpec{}
		}
	}
	query.Del(RootRequirementFileUnpackQueryKey)

	fingerprint := query.Get(RootRequirementFileFingerprintQueryKey)
	if fingerprint != "" {
		err, _, _ := fingerprintParse(fingerprint)
		if err != nil {
			return err, rootRequirementFileTargetSpec{}
		}
	}
	query.Del(RootRequirementFileFingerprintQueryKey)

	if len(query) > 0 {
		return fmt.Errorf("unsupported file requirement query parameter"), rootRequirementFileTargetSpec{}
	}

	return nil, rootRequirementFileTargetSpec{
		SourcePath:      filepath.Clean(linkPath),
		DestinationAs:   destinationAs,
		DestinationInto: destinationInto,
		Optional:        optional,
		Unpack:          unpack,
		Fingerprint:     fingerprint,
	}
}

func rootRequirementFileTargetString(sourcePath string, destinationAs string, destinationInto string, optional bool, unpack bool, fingerprint string) string {
	linkURL := url.URL{
		Scheme: "file",
		Opaque: filepath.ToSlash(filepath.Clean(sourcePath)),
	}
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
	if optional {
		query = append(query, RootRequirementFileOptionalQueryKey+"=true")
	}
	if unpack {
		query = append(query, RootRequirementFileUnpackQueryKey+"=true")
	}
	linkURL.RawQuery = strings.Join(query, "&")
	return linkURL.String()
}

func rootRequirementParseFileTarget(raw string) (error, rootRequirementFileTargetSpec, bool) {
	linkURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err, rootRequirementFileTargetSpec{}, false
	}
	if linkURL.Scheme != "file" {
		return nil, rootRequirementFileTargetSpec{}, false
	}

	err, fileSpec := rootRequirementFileTargetFromURL(linkURL)
	if err != nil {
		return err, rootRequirementFileTargetSpec{}, true
	}

	return nil, fileSpec, true
}

func RootRequirementFileTargetNormalize(raw string) (error, string) {
	err, fileSpec, isFile := rootRequirementParseFileTarget(raw)
	if err != nil {
		return err, ""
	}
	if !isFile {
		return fmt.Errorf("file requirement target must use file scheme: %s", raw), ""
	}

	return nil, rootRequirementFileTargetString(fileSpec.SourcePath, fileSpec.DestinationAs, fileSpec.DestinationInto, fileSpec.Optional, fileSpec.Unpack, fileSpec.Fingerprint)
}

func RootRequirementFileTargetString(sourcePath string, destinationAs string, destinationInto string, optional bool, unpack bool, fingerprint string) string {
	return rootRequirementFileTargetString(sourcePath, destinationAs, destinationInto, optional, unpack, fingerprint)
}

func rootRequirementFileIsWithin(basePath string, path string) (error, bool) {
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err, false
	}
	if relPath == "." || (relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator))) {
		return nil, true
	}
	return nil, false
}

func rootRequirementFileCopyFile(srcPath string, destPath string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, src)
	return err
}

func rootRequirementFileCopyDir(ctx *task.ExecutionContext, srcPath string, destPath string) error {
	srcPath, err := filepath.Abs(srcPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return err
	}

	shouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		if node.Info == nil {
			return nil, false
		}
		mode := node.Info.Mode()
		if mode&os.ModeSymlink == os.ModeSymlink {
			linkTarget, err := os.Readlink(node.Path)
			if err != nil {
				return err, false
			}
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(filepath.Dir(node.Path), linkTarget)
			}
			err, isInternal := rootRequirementFileIsWithin(srcPath, filepath.Clean(linkTarget))
			if err != nil {
				return err, false
			}
			return nil, !isInternal
		}
		if !node.Info.IsDir() {
			return nil, false
		}

		parentDir := filepath.Dir(node.VPath)
		err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
			BasePath: srcPath,
			Path:     parentDir,
		})
		if err != nil {
			return err, false
		}
		err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, true))
		if err != nil {
			return err, false
		}
		return nil, !match
	}

	onCopy := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		relPath, err := filepath.Rel(srcPath, node.VPath)
		if err != nil {
			return err, nil
		}
		if relPath == "." {
			return nil, nil
		}

		parentDir := filepath.Dir(node.VPath)
		err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
			BasePath: srcPath,
			Path:     parentDir,
		})
		if err != nil {
			return err, nil
		}
		err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, node.Info.IsDir()))
		if err != nil {
			return err, nil
		}
		if match {
			return nil, nil
		}

		dest := filepath.Join(destPath, relPath)
		mode := node.Info.Mode()
		switch {
		case mode.IsDir():
			return os.MkdirAll(dest, 0o755), nil
		case mode&os.ModeSymlink == os.ModeSymlink:
			linkTarget, err := os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}
			absLinkTarget := linkTarget
			if !filepath.IsAbs(absLinkTarget) {
				absLinkTarget = filepath.Join(filepath.Dir(node.Path), linkTarget)
			}
			err, isInternal := rootRequirementFileIsWithin(srcPath, filepath.Clean(absLinkTarget))
			if err != nil {
				return err, nil
			}
			if !isInternal {
				return nil, nil
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err, nil
			}
			return os.Symlink(linkTarget, dest), nil
		case mode.IsRegular():
			return rootRequirementFileCopyFile(node.Path, dest, mode), nil
		default:
			return fmt.Errorf("unsupported file requirement source type: %s", node.Path), nil
		}
	}

	err, _ = dydfs.Walk6(ctx, dydfs.Walk6Request{
		BasePath:   srcPath,
		Path:       srcPath,
		VPath:      srcPath,
		ShouldWalk: shouldWalk,
		OnPreMatch: onCopy,
	})
	return err
}

func rootRequirementFileImportSourceExact(ctx *task.ExecutionContext, srcPath string, destPath string) error {
	info, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}
	mode := info.Mode()
	if mode.IsDir() {
		return rootRequirementFileCopyDir(ctx, srcPath, destPath)
	}
	if mode&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(srcPath)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Join(filepath.Dir(srcPath), linkTarget)
		}
		return rootRequirementFileImportSourceExact(ctx, filepath.Clean(linkTarget), destPath)
	}
	if mode.IsRegular() {
		return rootRequirementFileCopyFile(srcPath, destPath, mode)
	}
	return fmt.Errorf("unsupported file requirement source type: %s", srcPath)
}

func rootRequirementFileArchiveDestinationName(srcPath string) string {
	base := filepath.Base(srcPath)
	for _, suffix := range []string{".tar.gz", ".tgz", ".tar"} {
		if strings.HasSuffix(base, suffix) {
			return strings.TrimSuffix(base, suffix)
		}
	}
	return base
}

func rootRequirementFileArchiveReader(srcPath string) (io.ReadCloser, error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(srcPath, ".tar") {
		return src, nil
	}
	if strings.HasSuffix(srcPath, ".tar.gz") || strings.HasSuffix(srcPath, ".tgz") {
		gz, err := gzip.NewReader(src)
		if err != nil {
			src.Close()
			return nil, err
		}
		return struct {
			io.Reader
			io.Closer
		}{Reader: gz, Closer: multiCloser{gz, src}}, nil
	}
	src.Close()
	return nil, fmt.Errorf("unsupported archive type for unpack=true: %s", srcPath)
}

type multiCloser []io.Closer

func (closers multiCloser) Close() error {
	var firstErr error
	for _, closer := range closers {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func rootRequirementFileSafeArchivePath(basePath string, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	destPath := filepath.Join(basePath, cleanName)
	err, isWithin := rootRequirementFileIsWithin(basePath, destPath)
	if err != nil {
		return "", err
	}
	if !isWithin {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	return destPath, nil
}

func rootRequirementFileValidateArchiveSymlink(basePath string, entryPath string, linkName string) error {
	if filepath.IsAbs(linkName) {
		return fmt.Errorf("unsafe absolute archive symlink target: %s", linkName)
	}
	linkPath := filepath.Clean(filepath.Join(filepath.Dir(entryPath), linkName))
	err, isWithin := rootRequirementFileIsWithin(basePath, linkPath)
	if err != nil {
		return err
	}
	if !isWithin {
		return fmt.Errorf("unsafe archive symlink target: %s", linkName)
	}
	return nil
}

func rootRequirementFileExtractTar(srcPath string, destPath string) error {
	reader, err := rootRequirementFileArchiveReader(srcPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		entryPath, err := rootRequirementFileSafeArchivePath(destPath, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(entryPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(entryPath), 0o755); err != nil {
				return err
			}
			dest, err := os.OpenFile(entryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(dest, tarReader)
			closeErr := dest.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(entryPath), 0o755); err != nil {
				return err
			}
			if err := rootRequirementFileValidateArchiveSymlink(destPath, entryPath, header.Linkname); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, entryPath); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive entry type for %s", header.Name)
		}
	}
	return nil
}

type RootRequirementFileBuildStemRequest struct {
	Garden          *SafeGardenReference
	SourcePath      string
	DestinationAs   string
	DestinationInto string
	Optional        bool
	Unpack          bool
}

func RootRequirementFileBuildStem(ctx *task.ExecutionContext, req RootRequirementFileBuildStemRequest) (error, *SafeHeapStemReference) {
	destinationAs := req.DestinationAs
	destinationInto := req.DestinationInto
	if destinationAs != "" && destinationInto != "" {
		return fmt.Errorf("file requirement cannot specify both as and into"), nil
	}
	if destinationAs != "" {
		var err error
		destinationAs, err = rootRequirementFilePlacementNormalize(destinationAs, true)
		if err != nil {
			return err, nil
		}
	} else {
		var err error
		destinationInto, err = rootRequirementFilePlacementNormalize(rootRequirementFilePlacementDefaultInto(destinationInto), false)
		if err != nil {
			return err, nil
		}
	}

	stemPath, err := os.MkdirTemp("", "dryad-file-*")
	if err != nil {
		return err, nil
	}
	defer os.RemoveAll(stemPath)

	if err := StemInit(stemPath); err != nil {
		return err, nil
	}
	destPath := ""
	if destinationAs != "" {
		destPath = filepath.Join(stemPath, destinationAs)
	} else if req.Unpack {
		destPath = filepath.Join(stemPath, destinationInto, rootRequirementFileArchiveDestinationName(req.SourcePath))
	} else {
		destPath = filepath.Join(stemPath, destinationInto, filepath.Base(req.SourcePath))
	}

	missingOptionalSource := false
	if req.Unpack {
		extractPath, err := os.MkdirTemp("", "dryad-file-unpack-*")
		if err != nil {
			return err, nil
		}
		defer os.RemoveAll(extractPath)
		if err := rootRequirementFileExtractTar(req.SourcePath, extractPath); err != nil {
			if req.Optional && os.IsNotExist(err) {
				missingOptionalSource = true
			} else {
				return err, nil
			}
		}
		if !missingOptionalSource {
			if err := rootRequirementFileCopyDir(ctx, extractPath, destPath); err != nil {
				return err, nil
			}
		}
	} else if err := rootRequirementFileImportSourceExact(ctx, req.SourcePath, destPath); err != nil {
		if req.Optional && os.IsNotExist(err) {
			missingOptionalSource = true
		} else {
			return err, nil
		}
	}

	err, _ = stemFinalize(ctx, stemPath)
	if err != nil {
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
	return stems.AddStem(ctx, HeapAddStemRequest{StemPath: stemPath})
}
