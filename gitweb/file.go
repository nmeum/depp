package gitweb

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/filemode"
)

// RepoFile represents information for a single file/blob.
type RepoFile struct {
	// TODO: Considering including objects.File?!
	mode filemode.FileMode
	Path string // Slash separated path
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

func (f *RepoFile) FilePath() string {
	return filepath.FromSlash(f.Path)
}

func (f *RepoFile) IsDir() bool {
	return f.mode == filemode.Dir
}

func (f *RepoFile) IsSubmodule() bool {
	return f.mode == filemode.Submodule
}

func (f *RepoFile) PathElements() []string {
	return strings.SplitN(f.Path, "/", -1)
}
