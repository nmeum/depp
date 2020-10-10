package main

import (
	git "github.com/libgit2/git2go"

	"os"
	"path"
	"path/filepath"
)

// Repo represents information required per repository.
type Repo struct {
	git  *git.Repository
	path string

	Title     string
	URL       string
	CurBranch string
	Readme    string
}

// RepoPage represents information required per reference.
type RepoPage struct {
	Repo

	tree   *git.Tree
	commit *git.Commit

	CurrentFile string
	IsDir       bool

	Commits []*git.Commit
}

func NewRepo(fp string) (*Repo, error) {
	var err error
	r := &Repo{path: fp}

	r.git, err = git.OpenRepository(fp)
	if err != nil {
		return nil, err
	}

	r.Title = filepath.Base(fp)
	r.URL = path.Join("git://git.8pit.net")                 // TODO
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

func (r *Repo) GetPage(ref *git.Reference, fp string) (*RepoPage, error) {
	var err error
	page := &RepoPage{Repo: *r}

	oid := ref.Target()
	page.commit, err = r.git.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	page.tree, err = page.commit.Tree()
	if err != nil {
		return nil, err
	}

	page.CurrentFile = fp
	if page.CurrentFile != "" {
		entry, err := page.tree.EntryByPath(fp)
		if err != nil {
			return nil, err
		}

		page.IsDir = entry.Type == git.ObjectTree
	} else {
		page.IsDir = true
	}

	// TODO: Make N configurable
	page.Commits, err = getCommits(page.commit, 5)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (r *RepoPage) FilesByRoot(prefix string) ([]os.FileInfo, error) {
	var ret error
	var entries []os.FileInfo
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != prefix {
			return 1 /* Skip passed entry */
		}

		fp := filepath.Join(r.path, root, e.Name)
		stat, err := os.Stat(fp) // TODO: Doesn't work for bare repos
		if err != nil {
			ret = err
		}

		entries = append(entries, stat)
		return 0
	})
	if ret != nil {
		return nil, ret
	}

	return entries, nil
}
