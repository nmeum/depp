package gitweb

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

var (
	ExpectedDirectory = errors.New("Expected directory")
	ExpectedSubmodule = errors.New("Expected submodule")
	ExpectedRegular   = errors.New("Expected regular file")
)

func (r *RepoPage) Files() ([]RepoFile, error) {
	if !r.CurrentFile.IsDir() {
		return nil, ExpectedDirectory
	}

	var entries []RepoFile
	basepath := filepath.Base(r.CurrentFile.Path)

	walker := object.NewTreeWalker(r.tree, false, nil)
	defer walker.Close()
	for {
		name, f, err := walker.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		relpath := filepath.Join(basepath, name)
		file := RepoFile{
			Path: filepath.ToSlash(relpath),
			mode: f.Mode,
		}

		entries = append(entries, file)
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

func (r *RepoPage) Blob() (*object.File, error) {
	if r.CurrentFile.IsDir() || r.CurrentFile.IsSubmodule() {
		return nil, ExpectedRegular
	}

	commit, err := r.Tip()
	if err != nil {
		return nil, err
	}
	return commit.File(r.CurrentFile.Path)
}

func (r *RepoPage) Submodule(file *RepoFile) (*object.File, error) {
	if !file.IsSubmodule() {
		return nil, ExpectedSubmodule
	}

	// git-go only seems to have very limited support for submodules
	// in bare repositories. Hence, just display .gitmodules for now.
	commit, err := r.Tip()
	if err != nil {
		return nil, err
	}
	return commit.File(".gitmodules")
}

func (r *RepoPage) findReadme() (string, error) {
	var result string

	walker := object.NewTreeWalker(r.tree, false, nil)
	defer walker.Close()
	for {
		_, f, err := walker.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if isReadme(f.Name) {
			result = f.Name
			break
		}
	}

	if result == "" {
		return "", os.ErrNotExist
	}

	return result, nil
}

func (r *RepoPage) Readme() (string, error) {
	if !r.CurrentFile.IsDir() {
		return "", ExpectedDirectory
	}

	fp, err := r.findReadme()
	if err != nil {
		return "", err
	}
	file, err := r.tree.File(fp)
	if err != nil {
		return "", err
	}

	return file.Contents()
}
