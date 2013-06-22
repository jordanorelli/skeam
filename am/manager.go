package am

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// an asset manager is responsible for managing references to files contained
// in source-controlled directories that associate to Go packages.  I suspect
// that there's some overlap here with the standard library go/build package.
type Manager struct {
	basePath    string // base path on the local filesystem
	baseURLPath string // base path for public-facing URLS
}

// given an import path for a Go package, builds an asset manager for that
// package.  Doesn't actually check if that package exists; this is kinda
// hazardous right now.  ¯\_(ツ)_/¯
func New(importPath, urlBase string) *Manager {
	return &Manager{basePath: pkgDir(importPath), baseURLPath: urlBase}
}

func (m *Manager) AbsPath(parts ...string) string {
	return filepath.Join(m.basePath, filepath.Join(parts...))
}

func (m *Manager) Open(parts ...string) (*os.File, error) {
	return os.Open(m.AbsPath(parts...))
}

func (m *Manager) URLPath(parts ...string) string {
	return path.Join(m.baseURLPath, path.Join(parts...))
}

func (m *Manager) ReadFile(parts ...string) ([]byte, error) {
	f, err := m.Open(parts...)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// getPathDirs gets the user's GOPATH environment variable and splits it along
// the OS-specific path separator, returning a slice of strings, one per path
// directory.
func getPathDirs() []string {
	path := os.Getenv("GOPATH")
	if path == "" {
		panic("am: GOPATH not set")
	}
	return strings.Split(path, string(os.PathListSeparator))
}

// getPkgDirCandidates gets the list of possible locations for a given import
// string.  The paths are not guaranteed to exist or to even be valid; results
// are derived from the user's $GOPATH environment variable, which is not
// necessarily clean.
func getPkgDirCandidates(importPath string) []string {
	pathDirs := getPathDirs()
	importParts := strings.Split(importPath, "/")
	candidates := make([]string, 0, len(pathDirs))
	for _, pathDir := range pathDirs {
		srcDir := filepath.Join(pathDir, "src")
		candidates = append(candidates, filepath.Join(srcDir, filepath.Join(importParts...)))
	}
	return candidates
}

// existingDirectories accepts a slice of strings and returns a slice of
// strings representing which of the given strings corresponds to an existing
// directory.  Note that this is prone to timing attacks, but it is presumed to
// not matter for this application; this may be an unsafe strategy to copy into
// other projects.
func existingDirectories(candidates []string) []string {
	dirs := make([]string, 0, len(candidates))
	for _, path := range candidates {
		if isDir(path) {
			dirs = append(dirs, path)
		}
	}
	return dirs
}

// isDir takes a file path and returns a boolean representing whether the path
// is or is not a valid directory.  Again, this is vulnerable to timing attacks
// and should be considered a Very Bad Idea in most contexts.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// getPkgDir takes an import path and returns a string representing the
// directory path of the package on the current machine.  This is generally a
// stupid thing to do, because we're generally not worried about the location
// of a package's source code files, since they may be out of sync with the
// actual binary, but in our case, we're using it to look up assets that have
// been made go-gettable.  If no package dir can be found, an empty string is
// returned.  The package may reasonably be installed into multiple workspaces.
// In this case, it's the first package found, as dictated by the user's
// $GOPATH environment variable.
func pkgDir(importPath string) string {
	dirs := existingDirectories(getPkgDirCandidates(importPath))
	if len(dirs) == 0 {
		return ""
	}
	return dirs[0]
}
