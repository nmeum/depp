package main

import (
	git "github.com/libgit2/git2go"

	"os"
	"path"
	"path/filepath"
)

// Repo represents information required per repository.
type Repo struct {
	git *git.Repository

	Title     string
	URL       string
	CurBranch string
	Readme    string
}

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	Commits []*git.Commit
	Tree    []os.FileInfo
}

func NewRepo(fp string) (*Repo, error) {
	var err error
	r := &Repo{}

	r.git, err = git.OpenRepository(fp)
	if err != nil {
		return nil, err
	}

	r.Title = filepath.Base(fp)
	r.URL = path.Join("git://git.8pit.net") // TODO
	r.Readme = "# Readme\n\nSomething something something." // TODO

	head, err := r.git.Head()
	if err != nil {
		return nil, err
	}

	r.CurBranch, err = head.Branch().Name()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repo) Branches() ([]string, error) {
	iterator, err := r.git.NewBranchIterator(git.BranchLocal)
	if err != nil {
		return []string{}, nil
	}

	var ret error
	var branches []string
	iterator.ForEach(func(b *git.Branch, t git.BranchType) error {
		name, err := b.Name()
		if err != nil {
			ret = err
		}

		branches = append(branches, name)
		return nil
	})
	if ret != nil {
		return []string{}, ret
	}

	return branches, nil
}

func (r *Repo) GetPage(ref *git.Reference) (*RepoPage, error) {
	page := &RepoPage{Repo: *r}

	oid := ref.Target()
	commit, err := r.git.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	// TODO: Make N configurable
	page.Commits, err = getCommits(commit, 5)
	if err != nil {
		return nil, err
	}

	return page, nil
}
