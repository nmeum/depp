package main

import (
	git "github.com/libgit2/git2go"

	"path"
	"path/filepath"
)

type RepoFile struct {
	Path  string /// Slash separated path
	IsDir bool
}

func (f *RepoFile) Name() string {
	return path.Base(f.Path)
}

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

func (r *Repo) Walk(fn func(*RepoPage) error) error {
	head, err := r.git.Head()
	if err != nil {
		return err
	}

	indexPage, err := r.GetPage(head, "")
	if err != nil {
		return err
	}

	var ret error
	indexPage.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root == "" {
			ret = fn(indexPage)
			if ret != nil {
				return -1
			}
		}

		fp := filepath.Join(root, e.Name)
		page, ret := r.GetPage(head, fp)
		if ret != nil {
			return -1
		}

		ret = fn(page)
		if ret != nil {
			return -1
		}

		return 0
	})

	return ret
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

		if page.IsDir {
			page.tree, err = r.git.LookupTree(entry.Id)
			if err != nil {
				panic(err)
				return nil, err
			}
		}
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

func (r *RepoPage) Files() ([]RepoFile, error) {
	var ret error
	var entries []RepoFile
	r.tree.Walk(func(root string, e *git.TreeEntry) int {
		if root != "" {
			return 1 /* Skip passed entry */
		}

		file := RepoFile{
			Path:  path.Join(r.CurrentFile, e.Name),
			IsDir: e.Type == git.ObjectTree,
		}

		entries = append(entries, file)
		return 0
	})
	if ret != nil {
		return nil, ret
	}

	return entries, nil
}

func (r *RepoPage) GetBlob(fp string) (string, error) {
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
