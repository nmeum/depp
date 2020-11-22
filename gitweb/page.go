package gitweb

import (
	git "github.com/libgit2/git2go"

	"errors"
	"path/filepath"
	"sort"
)

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	tree   *git.Tree
	commit *git.Commit

	CurrentFile RepoFile
}

func (r *RepoPage) Files() ([]RepoFile, error) {
	var ret error
	var entries []RepoFile
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != "" {
			return 1 /* Skip passed entry */
		}

		basepath := filepath.Base(r.CurrentFile.Path)
		relpath := filepath.Join(basepath, e.Name)

		file := RepoFile{
			Path: filepath.ToSlash(relpath),
			Type: e.Type,
		}

		entries = append(entries, file)
		return 0
	})
	if ret != nil {
		return nil, ret
	}

	sort.Sort(byType(entries))
	return entries, nil
}

func (r *RepoPage) Commits() ([]*git.Commit, error) {
	var i uint

	commit := r.commit
	commits := make([]*git.Commit, r.numCommits)

	for i = 0; i < r.numCommits; i++ {
		if commit == nil {
			break
		}

		commits[i] = commit
		commit = commit.Parent(0)
	}

	commits = commits[0:i] // Shrink to appropriate size
	return commits, nil

}

func (r *RepoPage) Blob(file *RepoFile) (string, error) {
	if file.Type != git.ObjectBlob {
		return "", errors.New("given RepoFile is not a blob")
	}
	fp := file.FilePath()

	entry, err := r.tree.EntryByPath(fp)
	if err != nil {
		return "", err
	}

	oid := entry.Id
	blob, err := r.git.LookupBlob(oid)
	if err != nil {
		return "", err
	}

	return string(blob.Contents()), nil
}

func (r *RepoPage) Submodule(file *RepoFile) (string, error) {
	// TODO: Submodules.Lookup does not work in bare repositories.
	// See: https://github.com/libgit2/libgit2/commit/477b3e047426d7ccddb6028416ff0fcc2541a0fd
	//
	// Therefore, this function simply returns the contents of .gitmodules.

	if !file.IsSubmodule() {
		return "", errors.New("given RepoFile is not a submodule")
	}

	entry, err := r.tree.EntryByPath(".gitmodules")
	if err != nil {
		return "", err
	}

	oid := entry.Id
	blob, err := r.git.LookupBlob(oid)
	if err != nil {
		return "", err
	}

	return string(blob.Contents()), nil
}
