package gitweb

import (
	git "github.com/libgit2/git2go/v30"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	tree   *git.Tree
	commit *git.Commit

	CurrentFile RepoFile
}

type CommitInfo struct {
	Commits []*git.Commit
	Total   uint
}

var readmeRegex = regexp.MustCompile("README|(README\\.[a-zA-Z0-9]+)")

func (r *RepoPage) Files() ([]RepoFile, error) {
	var ret error
	var entries []RepoFile
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != "" {
			return 1 // Skip passed entry
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

func (r *RepoPage) Commits() (*CommitInfo, error) {
	var oid git.Oid
	var total, numCommits uint

	walker, err := r.git.Walk()
	if err != nil {
		return nil, err
	}
	defer walker.Free()
	walker.Push(r.commit.AsObject().Id())

	commits := make([]*git.Commit, r.numCommits)
	for walker.Next(&oid) == nil {
		commit, err := r.git.LookupCommit(&oid)
		if err != nil {
			return nil, err
		}

		if numCommits < r.numCommits {
			commits[numCommits] = commit
			numCommits++
		}

		total++
	}

	commits = commits[0:numCommits] // Shrink to appropriate size
	return &CommitInfo{commits, total}, nil
}

func (r *RepoPage) Blob(file *RepoFile) ([]byte, error) {
	if file.Type != git.ObjectBlob {
		return []byte{}, errors.New("given RepoFile is not a blob")
	}
	fp := file.FilePath()

	entry, err := r.tree.EntryByPath(fp)
	if err != nil {
		return []byte{}, err
	}

	oid := entry.Id
	blob, err := r.git.LookupBlob(oid)
	if err != nil {
		return []byte{}, err
	}

	return blob.Contents(), nil
}

func (r *RepoPage) Submodule(file *RepoFile) ([]byte, error) {
	if !file.IsSubmodule() {
		return []byte{}, errors.New("given RepoFile is not a submodule")
	}
	fp := file.FilePath()

	submodule, err := r.git.Submodules.Lookup(fp)
	if git.IsErrorClass(err, git.ErrClassSubmodule) {
		// TODO: Submodules.Lookup does not work in bare repositories.
		// See: https://github.com/libgit2/libgit2/commit/477b3e047426d7ccddb6028416ff0fcc2541a0fd
		gitmodules := &RepoFile{".gitmodules", git.ObjectBlob}
		return r.Blob(gitmodules)
	}

	out := fmt.Sprintf("%v @ %v", submodule.Url(), submodule.IndexId())
	return []byte(out), nil
}

func (r *RepoPage) matchFile(reg *regexp.Regexp) *git.TreeEntry {
	var result *git.TreeEntry
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != "" {
			return 1 // Different directory
		}

		if e.Type == git.ObjectBlob && reg.MatchString(e.Name) {
			result = e
			return -1 // Stop the walk
		}

		return 0
	})

	return result
}

func (r *RepoPage) Readme() (string, error) {
	entry := r.matchFile(readmeRegex)
	if entry == nil {
		return "", os.ErrNotExist
	}

	blob, err := r.git.LookupBlob(entry.Id)
	if err != nil {
		return "", err
	}

	return string(blob.Contents()), nil
}
