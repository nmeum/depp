package gitweb

import (
	git "github.com/libgit2/git2go/v30"

	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Repo represents information required per repository.
type Repo struct {
	git        *git.Repository
	maxCommits uint

	Path  string
	Title string
	URL   string
}

// File name of the git description file.
const descFn = "description"

func NewRepo(fp string, gitServer *url.URL, commits uint) (*Repo, error) {
	var err error

	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	r := &Repo{Path: absFp}

	r.git, err = git.OpenRepository(absFp)
	if err != nil {
		return nil, err
	}

	r.Title = filepath.Base(absFp)
	if gitServer != nil {
		r.URL = gitServer.String()
	}

	ext := strings.LastIndex(r.Title, ".git")
	if ext > 0 {
		r.Title = r.Title[0:ext]
	}

	r.maxCommits = commits
	return r, nil
}

func (r *Repo) Tip() (*git.Commit, error) {
	head, err := r.git.Head()
	if err != nil {
		return nil, err
	}

	oid := head.Target()
	commit, err := r.git.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (r *Repo) Walk(fn func(*RepoPage) error) error {
	head, err := r.git.Head()
	if err != nil {
		return err
	}

	indexPage, err := r.Page(head, "")
	if err != nil {
		return err
	}

	// . is not included by tree.Walk()
	err = fn(indexPage)
	if err != nil {
		return err
	}

	var ret error
	indexPage.tree.Walk(func(root string, e *git.TreeEntry) int {
		fp := filepath.Join(root, e.Name)
		page, err := r.Page(head, fp)
		if err != nil {
			ret = err
			return -1
		}

		ret = fn(page)
		if err != nil {
			ret = err
			return -1
		}

		return 0
	})

	return ret
}

func (r *Repo) Page(ref *git.Reference, fp string) (*RepoPage, error) {
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

	// TODO: Find out how to retrieve the TreeEntry for /
	page.CurrentFile = RepoFile{Path: filepath.ToSlash(fp)}
	if page.CurrentFile.Path != "" {
		entry, err := page.tree.EntryByPath(fp)
		if err != nil {
			return nil, err
		}
		page.CurrentFile.Type = entry.Type

		if page.CurrentFile.IsDir() {
			page.tree, err = r.git.LookupTree(entry.Id)
			if err != nil {
				panic(err)
				return nil, err
			}
		}
	} else {
		page.CurrentFile.Type = git.ObjectTree
	}

	return page, nil
}

func (r *Repo) Description() (string, error) {
	fp := filepath.Join(r.Path, descFn)

	desc, err := os.ReadFile(fp)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return string(desc), nil
}
