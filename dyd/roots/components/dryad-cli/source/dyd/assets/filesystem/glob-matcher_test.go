package fs2

import (
    "testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobPattern_BasicTable(t *testing.T) {
	assert := assert.New(t)

	type tc struct {
		name    string
		base    string
		pattern string
		path    string
		expected    GlobPatternMatchResult
	}

	cases := []tc{
		// --- literals ---
		{name: "literal-match", base: "", pattern: "foo", path: "/foo", expected: PATTERN_INCLUDE},
		{name: "literal-superstring", base: "", pattern: "foo", path: "/foo2", expected: PATTERN_NO_MATCH},
		{name: "literal-prefix", base: "", pattern: "foo", path: "/fo", expected: PATTERN_NO_MATCH},

		// --- '*' wildcard ---
		{name: "star-root", base: "", pattern: "*.txt", path: "/a.txt", expected: PATTERN_INCLUDE},
		{name: "star-deep", base: "", pattern: "*.txt", path: "/a/b/c/file.txt", expected: PATTERN_INCLUDE},
		{name: "star-nonmatching-ext", base: "", pattern: "*.txt", path: "/a/b/c/file.txtx", expected: PATTERN_NO_MATCH},

		// --- '?' single-char ---
		{name: "question-one-char", base: "", pattern: "file.?", path: "/file.a", expected: PATTERN_INCLUDE},
		{name: "question-two-chars", base: "", pattern: "file.?", path: "/file.ab", expected: PATTERN_NO_MATCH},
		{name: "question-zero-chars", base: "", pattern: "file.?", path: "/file.", expected: PATTERN_NO_MATCH},

		// --- '[]' char class ---
		{name: "class-a", base: "", pattern: "[ab].go", path: "/a.go", expected: PATTERN_INCLUDE},
		{name: "class-b", base: "", pattern: "[ab].go", path: "/b.go", expected: PATTERN_INCLUDE},
		{name: "class-miss", base: "", pattern: "[ab].go", path: "/c.go", expected: PATTERN_NO_MATCH},

		// --- '**' wildcard ---
		{name: "globstar-prefix", base: "", pattern: "**/foo", path: "/x/y/foo", expected: PATTERN_INCLUDE},
		{name: "globstar-suffix", base: "", pattern: "foo/**", path: "/foo/x/y", expected: PATTERN_INCLUDE},

		{name: "normalized-slashes", base: "", pattern: "bar/**/*.go", path: `\bar\pkg\file.go`, expected: PATTERN_INCLUDE},		
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// compile pattern
			err, pattern := NewGlobPattern(tc.pattern, tc.base)
			assert.Nil(err)
			assert.NotNil(pattern)

			path := NewGlobPath(tc.path, false)

			err, match := pattern.Match(path)
			assert.Nil(err)
			assert.Equal(tc.expected, match, "pattern=%q base=%q path=%q -> %q compiled=%q", tc.pattern, tc.base, tc.path, path.path, pattern.pattern)
		})
	}
}


func TestGlobMatcher0(t *testing.T) {
	assert := assert.New(t)

	type tc struct {
		name    string
		path    string
		expected    bool
	}

	cases := []tc{
		// literals
		{name: "exact-root-file", path: "/foo", expected: true},
		{name: "literal-superstring", path: "/foo2", expected: false},
		{name: "literal-prefix", path: "/fo", expected: false},
		{name: "no-root", path: "foo", expected: true},

		// nested paths (unanchored -> any depth)
		{name: "subdir-file", path: "/a/b/foo", expected: true},
		{name: "deeply-nested", path: "/x/y/z/foo", expected: true},

		// non-matches by segment
		{name: "segment-followed-by-more", path: "/foo/bar", expected: false},
		{name: "extension-differs", path: "/foo.txt", expected: false},

		// normalization cases
		{name: "double-slashes", path: "//foo", expected: true},
		{name: "dot-prefix", path: "./foo", expected: true},
		{name: "dotdot-normalization", path: "./x/../foo", expected: true},
		{name: "windows-separators-match", path: `\a\foo`, expected: true},
		{name: "windows-separators-nonmatch", path: `\a\foo2`, expected: false},

		// case-sensitivity
		{name: "case-sensitive-miss", path: "/Foo", expected: false},

		// root path should not match "foo"
		{name: "root-path", path: "/", expected: false},
	}

	err, matcher := NewGlobMatcher([]string{"", "# bar", "foo", }, "", nil)
	assert.Nil(err)
	assert.NotNil(matcher)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := NewGlobPath(tc.path, false)

			err, match := matcher.Match(path)
			assert.Nil(err)
			assert.Equal(tc.expected, match, "path=%q -> %q", tc.path, path.path)
		})
	}
}

func TestGlobMatcher1(t *testing.T) {
	assert := assert.New(t)

	type tc struct {
		name    string
		path    string
		expected    bool
	}

	cases := []tc{
		// literals
		{name: "exact-root-file", path: "/foo", expected: false},
		{name: "literal-superstring", path: "/foo2", expected: false},
		{name: "literal-prefix", path: "/fo", expected: false},
		{name: "no-root", path: "foo", expected: false},

		// nested paths (unanchored -> any depth)
		{name: "subdir-file", path: "/a/b/foo", expected: false},
		{name: "deeply-nested", path: "/x/y/z/foo", expected: false},

		// non-matches by segment
		{name: "segment-followed-by-more", path: "/foo/bar", expected: false},
		{name: "extension-differs", path: "/foo.txt", expected: false},

		// normalization cases
		{name: "double-slashes", path: "//foo", expected: false},
		{name: "dot-prefix", path: "./foo", expected: false},
		{name: "dotdot-normalization", path: "./x/../foo", expected: false},
		{name: "windows-separators-match", path: `\a\foo`, expected: false},
		{name: "windows-separators-nonmatch", path: `\a\foo2`, expected: false},

		// case-sensitivity
		{name: "case-sensitive-miss", path: "/Foo", expected: false},

		// root path should not match "foo"
		{name: "root-path", path: "/", expected: false},
	}

	err, matcher := NewGlobMatcher([]string{}, "", nil)
	assert.Nil(err)
	assert.NotNil(matcher)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := NewGlobPath(tc.path, false)

			err, match := matcher.Match(path)
			assert.Nil(err)
			assert.Equal(tc.expected, match, "path=%q -> %q", tc.path, path.path)
		})
	}
}

func TestGlobMatcher2(t *testing.T) {
	assert := assert.New(t)

	type tc struct {
		name    string
		path    string
		expected    bool
	}

	cases := []tc{
		// literals
		{name: "exact-root-file", path: "/foo", expected: true},
		{name: "literal-superstring", path: "/foo2", expected: false},
		{name: "literal-prefix", path: "/fo", expected: false},
		{name: "no-root", path: "foo", expected: true},

		// nested paths (unanchored -> any depth)
		{name: "subdir-file", path: "/a/b/foo", expected: true},
		{name: "deeply-nested", path: "/x/y/z/foo", expected: true},

		// non-matches by segment
		{name: "segment-followed-by-more", path: "/foo/bar", expected: false},
		{name: "extension-differs", path: "/foo.txt", expected: false},

		// normalization cases
		{name: "double-slashes", path: "//foo", expected: true},
		{name: "dot-prefix", path: "./foo", expected: true},
		{name: "dotdot-normalization", path: "./x/../foo", expected: true},
		{name: "windows-separators-match", path: `\a\foo`, expected: true},
		{name: "windows-separators-nonmatch", path: `\a\foo2`, expected: false},

		// case-sensitivity
		{name: "case-sensitive-miss", path: "/Foo", expected: false},

		// root path should not match "foo"
		{name: "root-path", path: "/", expected: false},
	}

	err, matcher := NewGlobMatcher([]string{"", "# bar", "foo", }, "", nil)
	assert.Nil(err)
	assert.NotNil(matcher)

	err, matcher = NewGlobMatcher([]string{}, "", matcher)
	assert.Nil(err)
	assert.NotNil(matcher)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := NewGlobPath(tc.path, false)

			err, match := matcher.Match(path)
			assert.Nil(err)
			assert.Equal(tc.expected, match, "path=%q -> %q", tc.path, path.path)
		})
	}
}