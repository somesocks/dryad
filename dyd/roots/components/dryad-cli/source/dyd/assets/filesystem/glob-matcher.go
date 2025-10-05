package fs2

import (
	"errors"

	"path"
	"path/filepath"
    "strings"

	"os"
	"io/fs"

	"github.com/bmatcuk/doublestar/v4"
)


type GlobPath struct {
	path string
	is_dir bool
}

func NewGlobPath(raw_path string, is_dir bool) (GlobPath) {
	raw_path = strings.ReplaceAll(raw_path, "\\", "/")
	raw_path = path.Clean(raw_path)
	if !strings.HasPrefix(raw_path, "/") {
		raw_path = "/" + raw_path
	}
	return GlobPath{path: raw_path, is_dir: is_dir}
}

type GlobPatternMatchResult int

const (
	PATTERN_NO_MATCH GlobPatternMatchResult = iota
	PATTERN_INCLUDE
	PATTERN_EXCLUDE
)

type GlobPattern struct {
	pattern string
	inclusion bool
	matches_dirs bool
	matches_files bool
}

func (p * GlobPattern) Match(path GlobPath) (error, GlobPatternMatchResult) {
	can_match := (path.is_dir && p.matches_dirs) || (!path.is_dir && p.matches_files)
	if (!can_match) {
		return nil, PATTERN_NO_MATCH
	}

	matched, err := doublestar.Match(p.pattern, path.path)
	if err != nil {
		return err, PATTERN_NO_MATCH
	} else if !matched {
		return nil, PATTERN_NO_MATCH
	} else if p.inclusion {
		return nil, PATTERN_INCLUDE
	} else {
		return nil, PATTERN_EXCLUDE
	}
}

func NewGlobPattern (pattern string, base string) (error, *GlobPattern) {
	inclusion := true
	if pattern == "" || strings.TrimSpace(pattern) == "" {
		// empty line -> no pattern
		return nil, nil
	} else if strings.HasPrefix(pattern, "#")  {
		// comment line -> no pattern
		return nil, nil
	} else if strings.HasPrefix(pattern, "!") {
		inclusion = false
		pattern = pattern[1:]
	} else if strings.HasPrefix(pattern, "\\!") || strings.HasPrefix(pattern, "\\#") {		
		pattern = pattern[1:]
	}

	anchored := strings.HasPrefix(pattern, "/")
	directory := strings.HasSuffix(pattern, "/")

	if anchored {
		pattern = base + "/" + pattern[1:]
	} else {
		pattern = base + "/**/" + pattern
	}

	pattern = strings.ReplaceAll(pattern, "\\", "/")
	pattern = path.Clean(pattern)

	matches_dirs := true
	matches_files := !directory

	err := doublestar.ValidatePattern(pattern)	
	if !err {
		return errors.New("ParseGlobPattern: invalid glob pattern " + pattern), nil
	}
	
	return nil, &GlobPattern{
		pattern:  pattern,
		inclusion: inclusion,
		matches_dirs: matches_dirs,
		matches_files: matches_files,
	}
}

type GlobMatcher struct {
	patterns []GlobPattern
	parent *GlobMatcher
}

func (m * GlobMatcher) Match(path GlobPath) (error, bool) {
	for _, pattern := range m.patterns {
		err, result := pattern.Match(path)
		if err != nil {
			return err, false
		}
		switch result {
		case PATTERN_EXCLUDE:
			return nil, false
		case PATTERN_INCLUDE:
			return nil, true
		case PATTERN_NO_MATCH:
		}
	}

	if m.parent != nil {
		err, res := m.parent.Match(path)
		return err, res
	} else {
		return nil, false
	}
}

func NewGlobMatcher(rules []string, base string, parent *GlobMatcher) (error, *GlobMatcher) {
	patterns := []GlobPattern{}

	// reverse order of rules to implement "last rule wins"
	for i := len(rules) - 1; i >= 0; i-- {
		err, pattern := NewGlobPattern(rules[i], base)
		if err != nil {
			return err, nil
		}
		if pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}

	return nil, &GlobMatcher{
		patterns: patterns,
		parent: parent,
	}
}

func NewGlobMatcherFromFile(filePath string, parent *GlobMatcher) (error, *GlobMatcher) {
	base := filepath.Dir(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Treat "file not found" as an empty matcher
		if errors.Is(err, fs.ErrNotExist) {
			return nil, &GlobMatcher{parent: parent}
		}
		return err, nil
	}

	// Normalize newlines and strip BOM
	s := string(data)
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	rules := strings.Split(s, "\n")
	return NewGlobMatcher(rules, base, parent)
}

