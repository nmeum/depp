package gitweb

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/ioutil"
)

type RepoPage struct {
	*Repo

	// Underlying file which is present on this page.
	CurrentFile RepoFile

	// Tree, if the underlying file is a directory.
	tree *object.Tree
}

type CommitInfo struct {
	Commits []*object.Commit
	Total   uint
}

var readmeRegex = regexp.MustCompile(`README|(README\.[a-zA-Z0-9]+)`)

func (r *RepoPage) Files() ([]RepoFile, error) {
	if !r.CurrentFile.IsDir {
		return nil, nil
	}
	basepath := filepath.Base(r.CurrentFile.Path)

	var entries []RepoFile
	err := r.tree.Files().ForEach(func(f *object.File) error {
		relpath := filepath.Join(basepath, f.Name)
		file := RepoFile{
			Path:  filepath.ToSlash(relpath),
			IsDir: f.Mode == filemode.Dir,
		}

		entries = append(entries, file)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Sort(byType(entries))
	return entries, nil
}

func (r *RepoPage) Commits() (*CommitInfo, error) {
	var total, numCommits uint

	logOpts := &git.LogOptions{}
	if r.CurrentFile.Path != "" {
		logOpts.PathFilter = func(fp string) bool {
			return fp == r.CurrentFile.Path
		}
	}
	iter, err := r.git.Log(logOpts)
	if err != nil {
		return nil, err
	}

	commits := make([]*object.Commit, r.maxCommits)
	err = iter.ForEach(func(c *object.Commit) error {
		if numCommits < r.maxCommits {
			commits[numCommits] = c
			numCommits++
		}

		total++
		return nil
	})
	if err != nil {
		return nil, err
	}

	commits = commits[0:numCommits] // Shrink to appropriate size
	return &CommitInfo{commits, total}, nil
}

// TODO: Provide a better API which allows us to use File.IsBinary etc
func (r *RepoPage) Blob() ([]byte, error) {
	if r.CurrentFile.IsDir {
		return []byte{}, nil
	}

	commit, err := r.Tip()
	if err != nil {
		return nil, err
	}

	f, err := commit.File(r.CurrentFile.Path)
	if err != nil {
		return nil, err
	}
	reader, err := f.Reader()
	if err != nil {
		return nil, err
	}
	defer ioutil.CheckClose(reader, &err)

	return io.ReadAll(reader)
}

func (r *RepoPage) Submodule(file *RepoFile) ([]byte, error) {
	return []byte{}, errors.New("submodule support not yet implemented")
}

func (r *RepoPage) matchFile(reg *regexp.Regexp) *object.File {
	var result *object.File
	err := r.tree.Files().ForEach(func(f *object.File) error {
		if reg.MatchString(f.Name) {
			result = f
			return errors.New("found match") // stop iter
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return result
}

func (r *RepoPage) Readme() (string, error) {
	if !r.CurrentFile.IsDir {
		return "", nil
	}

	entry := r.matchFile(readmeRegex)
	if entry == nil {
		return "", os.ErrNotExist
	}

	return entry.Contents()
}
