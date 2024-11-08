package gitweb

import (
	"path"
	"path/filepath"
	"strings"
)

// RepoFile represents information for a single file/blob.
type RepoFile struct {
	// TODO: Considering including objects.File?!
	IsDir bool   // TODO: could store plumbing/filemode here
	Path  string // Slash separated path
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

func (f *RepoFile) FilePath() string {
	return filepath.FromSlash(f.Path)
}

func (f *RepoFile) IsSubmodule() bool {
	// TODO
	return false
}

func (f *RepoFile) PathElements() []string {
	return strings.SplitN(f.Path, "/", -1)
}
