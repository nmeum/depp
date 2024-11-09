package gitweb

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

	var numEntries int
	entries := make([]RepoFile, len(r.tree.Entries))

	seen := make(map[string]bool)
	err := r.tree.Files().ForEach(func(f *object.File) error {
		name := f.Name
		isDir := f.Mode == filemode.Dir

		slash := strings.IndexByte(f.Name, '/')
		if slash != -1 {
			name = f.Name[0:slash]
			isDir = true
			if seen[name] {
				return nil
			} else {
				seen[name] = true
			}
		}

		relpath := filepath.Join(basepath, name)
		file := RepoFile{
			Path:  filepath.ToSlash(relpath),
			IsDir: isDir,
		}

		entries[numEntries] = file
		numEntries++

		return nil
	})
	if err != nil {
		return nil, err
	}

	entries = entries[0:numEntries]
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

func (r *RepoPage) matchFile(reg *regexp.Regexp) (*object.File, error) {
	foundErr := errors.New("found match")

	var result *object.File
	err := r.tree.Files().ForEach(func(f *object.File) error {
		// Do not search in subdirectory of the directory.
		if strings.IndexByte(f.Name, filepath.Separator) != -1 {
			return nil
		}

		if reg.MatchString(f.Name) {
			result = f
			return foundErr // stop iter
		}

		return nil
	})
	if err == foundErr {
		return result, nil
	} else if err != nil {
		return nil, err
	}

	return nil, os.ErrNotExist
}

func (r *RepoPage) Readme() (string, error) {
	if !r.CurrentFile.IsDir {
		return "", nil
	}

	entry, err := r.matchFile(readmeRegex)
	if err != nil {
		return "", err
	}

	return entry.Contents()
}
