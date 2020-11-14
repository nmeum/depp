package gitweb

import (
	"path"
	"path/filepath"
	"strings"

	git "github.com/libgit2/git2go"
)

// RepoFile represents information for a single file/blob.
type RepoFile struct {
	Path string // Slash separated path
	Type git.ObjectType
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

func (f *RepoFile) FilePath() string {
	return filepath.FromSlash(f.Path)
}

func (f *RepoFile) IsDir() bool {
	return f.Type == git.ObjectTree
}

func (f *RepoFile) IsSubmodule() bool {
	return f.Type == git.ObjectCommit
}

func (f *RepoFile) PathElements() []string {
	return strings.SplitN(f.Path, "/", -1)
}
