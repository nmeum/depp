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

var readmeRegex = regexp.MustCompile(`README|(README\.[a-zA-Z0-9]+)`)

var (
	expectedDirectory = errors.New("expected directory")
	expectedSubmodule = errors.New("expected submodule")
	expectedRegular   = errors.New("expected regular file")
)

func (r *RepoPage) Files() ([]RepoFile, error) {
	if !r.CurrentFile.IsDir() {
		return nil, expectedDirectory
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
		return nil, expectedRegular
	}

	commit, err := r.Tip()
	if err != nil {
		return nil, err
	}
	return commit.File(r.CurrentFile.Path)
}

func (r *RepoPage) Submodule(file *RepoFile) (*object.File, error) {
	if !file.IsSubmodule() {
		return nil, expectedSubmodule
	}

	// git-go only seems to have very limited support for submodules
	// in bare repositories. Hence, just display .gitmodules for now.
	commit, err := r.Tip()
	if err != nil {
		return nil, err
	}
	return commit.File(".gitmodules")
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
	if !r.CurrentFile.IsDir() {
		return "", expectedDirectory
	}

	entry, err := r.matchFile(readmeRegex)
	if err != nil {
		return "", err
	}

	return entry.Contents()
}
