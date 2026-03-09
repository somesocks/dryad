package filepath

import (
	"dryad/diagnostics"
	"io/fs"
	stdfilepath "path/filepath"
)

const Separator = stdfilepath.Separator

type WalkFunc = stdfilepath.WalkFunc

var Base = stdfilepath.Base
var Clean = stdfilepath.Clean
var Dir = stdfilepath.Dir
var IsAbs = stdfilepath.IsAbs
var Join = stdfilepath.Join
var Split = stdfilepath.Split
var ToSlash = stdfilepath.ToSlash

var abs = diagnostics.BindA1R1(
	"filepath.abs",
	func(path string) string {
		return path
	},
	func(path string) (error, string) {
		resolvedPath, err := stdfilepath.Abs(path)
		return err, resolvedPath
	},
)

func Abs(path string) (string, error) {
	err, resolvedPath := abs(path)
	return resolvedPath, err
}

var evalSymlinks = diagnostics.BindA1R1(
	"filepath.eval_symlinks",
	func(path string) string {
		return path
	},
	func(path string) (error, string) {
		resolvedPath, err := stdfilepath.EvalSymlinks(path)
		return err, resolvedPath
	},
)

func EvalSymlinks(path string) (string, error) {
	err, resolvedPath := evalSymlinks(path)
	return resolvedPath, err
}

var glob = diagnostics.BindA1R1(
	"filepath.glob",
	func(pattern string) string {
		return pattern
	},
	func(pattern string) (error, []string) {
		matches, err := stdfilepath.Glob(pattern)
		return err, matches
	},
)

func Glob(pattern string) ([]string, error) {
	err, matches := glob(pattern)
	return matches, err
}

var rel = diagnostics.BindA2R1(
	"filepath.rel",
	func(basePath string, targetPath string) string {
		return targetPath
	},
	func(basePath string, targetPath string) (error, string) {
		relPath, err := stdfilepath.Rel(basePath, targetPath)
		return err, relPath
	},
)

func Rel(basePath string, targetPath string) (string, error) {
	err, relPath := rel(basePath, targetPath)
	return relPath, err
}

var WalkDir = diagnostics.BindA2R0(
	"filepath.walk_dir",
	func(root string, _ fs.WalkDirFunc) string {
		return root
	},
	stdfilepath.WalkDir,
)
